package app

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/subtle"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/mguhlin/pagespindle/internal/book"
	"github.com/mguhlin/pagespindle/internal/engine"
	"io"
	"io/fs"
	"mime"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"
)

//go:embed web/* web/assets/*
var assets embed.FS

type Job struct {
	ID       int            `json:"id"`
	State    string         `json:"state"`
	Message  string         `json:"message"`
	Done     int            `json:"done"`
	Total    int            `json:"total"`
	Document *book.Document `json:"document,omitempty"`
	Files    []string       `json:"files"`
	paths    map[string]string
	dir      string
	cancel   context.CancelFunc
	active   bool
}
type Server struct {
	Key, Host string
	Engine    *engine.Engine
	mu        sync.Mutex
	job       *Job
	next      int
	worker    sync.WaitGroup
	clearing  bool
}

func New(key, host, root string) *Server {
	return &Server{Key: key, Host: host, Engine: engine.New(root)}
}
func (s *Server) Handler() http.Handler {
	web, _ := fs.Sub(assets, "web")
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.FS(web)))
	mux.HandleFunc("/api/status", func(w http.ResponseWriter, r *http.Request) { s.json(w, s.Engine.Status()) })
	mux.HandleFunc("/api/new", s.newDocument)
	mux.HandleFunc("/api/import", s.importDocument)
	mux.HandleFunc("/api/prepare", s.prepare)
	mux.HandleFunc("/api/render", s.render)
	mux.HandleFunc("/api/job", s.getJob)
	mux.HandleFunc("/api/cancel", s.cancel)
	mux.HandleFunc("/api/clear", s.clear)
	mux.HandleFunc("/api/download", s.download)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; img-src 'self' data:; connect-src 'self'; media-src 'self' blob:; object-src 'none'; base-uri 'none'; form-action 'self'; frame-ancestors 'none'")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("Referrer-Policy", "no-referrer")
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
		if r.Host != s.Host {
			http.Error(w, "Wrong local host", 403)
			return
		}
		if origin := r.Header.Get("Origin"); origin != "" && origin != "http://"+s.Host {
			http.Error(w, "Cross-origin request denied", 403)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/api/") {
			if subtle.ConstantTimeCompare([]byte(r.Header.Get("X-PageSpindle-Key")), []byte(s.Key)) != 1 {
				http.Error(w, "Open PageSpindle from its launcher to start a session", 403)
				return
			}
			if r.Method != "POST" && r.URL.Path != "/api/status" && r.URL.Path != "/api/job" {
				http.Error(w, "Use POST", 405)
				return
			}
			if r.Method == "POST" && r.Header.Get("Origin") != "http://"+s.Host {
				http.Error(w, "Missing local origin", 403)
				return
			}
			r.Body = http.MaxBytesReader(w, r.Body, book.MaxInput+65536)
		}
		mux.ServeHTTP(w, r)
	})
}
func (s *Server) json(w http.ResponseWriter, v any) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(v)
}
func (s *Server) fail(w http.ResponseWriter, e error) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(400)
	json.NewEncoder(w).Encode(map[string]string{"error": e.Error()})
}
func decode(r *http.Request, v any) error {
	d := json.NewDecoder(r.Body)
	d.DisallowUnknownFields()
	if e := d.Decode(v); e != nil {
		return errors.New("The request is invalid or too large")
	}
	if d.Decode(&struct{}{}) != io.EOF {
		return errors.New("Unexpected request content")
	}
	return nil
}
func (s *Server) newDocument(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Text  string `json:"text"`
		Title string `json:"title"`
	}
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	d, e := book.New(req.Text, req.Title)
	if e != nil {
		s.fail(w, e)
		return
	}
	s.json(w, d)
}
func (s *Server) importDocument(w http.ResponseWriter, r *http.Request) {
	if e := r.ParseMultipartForm(1 << 20); e != nil {
		s.fail(w, errors.New("Choose a file up to 16 MiB"))
		return
	}
	defer r.MultipartForm.RemoveAll()
	f, h, e := r.FormFile("file")
	if e != nil {
		s.fail(w, errors.New("Choose a text, Markdown, EPUB, PDF, or PageSpindle project file"))
		return
	}
	defer f.Close()
	b, e := io.ReadAll(io.LimitReader(f, book.MaxInput+1))
	if e != nil || len(b) > book.MaxInput {
		s.fail(w, errors.New("File exceeds 16 MiB"))
		return
	}
	title := strings.TrimSuffix(filepath.Base(h.Filename), filepath.Ext(h.Filename))
	var text string
	switch strings.ToLower(filepath.Ext(h.Filename)) {
	case ".json":
		d, e := book.Load(b)
		if e != nil {
			s.fail(w, e)
			return
		}
		s.json(w, d)
		return
	case ".txt", ".md":
		text = string(b)
	case ".epub":
		text, e = book.EPUB(b)
	case ".pdf":
		ctx, cancel := context.WithTimeout(r.Context(), 20*time.Second)
		defer cancel()
		exe, err := os.Executable()
		if err != nil {
			s.fail(w, err)
			return
		}
		c := exec.CommandContext(ctx, exe, "--extract-pdf")
		c.Env = append(os.Environ(), "GOMEMLIMIT=256MiB")
		c.Stdin = bytes.NewReader(b)
		var out bytes.Buffer
		c.Stdout = &limitedWriter{W: &out, N: book.MaxText + 1}
		c.Stderr = io.Discard
		e = c.Run()
		text = out.String()
		if e != nil {
			e = errors.New("PDF extraction failed or timed out. Try a smaller text-based PDF; scanned pages need OCR first")
		}
	default:
		e = errors.New("Supported files: TXT, Markdown, EPUB, searchable PDF, and PageSpindle project JSON")
	}
	if e != nil {
		s.fail(w, e)
		return
	}
	d, e := book.New(text, title)
	if e != nil {
		s.fail(w, errors.New("No usable text was extracted. Scans need OCR; review PDF/EPUB reading order before narration"))
		return
	}
	s.json(w, d)
}

type limitedWriter struct {
	W io.Writer
	N int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if len(p) > l.N {
		return 0, errors.New("Output limit")
	}
	n, e := l.W.Write(p)
	l.N -= n
	return n, e
}
func (s *Server) begin(kind string, total int) (*Job, context.Context, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.clearing || (s.job != nil && s.job.active) {
		return nil, nil, errors.New("Wait for the current operation or cancel it first")
	}
	if s.job != nil && s.job.dir != "" {
		if err := os.RemoveAll(s.job.dir); err != nil {
			return nil, nil, errors.New("Previous temporary audio could not be cleared; close PageSpindle and retry")
		}
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.next++
	j := &Job{ID: s.next, State: "running", Message: kind, Total: total, cancel: cancel, active: true, paths: map[string]string{}, Files: []string{}}
	s.job = j
	s.worker.Add(1)
	return j, ctx, nil
}
func (s *Server) update(j *Job, done int, message string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == j && j.State == "running" {
		j.Done = done
		j.Message = message
	}
}
func (s *Server) finish(j *Job, e error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	j.active = false
	if s.job != j {
		return
	}
	if j.State == "cancelled" || j.State == "cancelling" {
		j.State = "cancelled"
		j.Message = "Cancelled. Your text is unchanged."
		if j.dir != "" {
			os.RemoveAll(j.dir)
		}
		return
	}
	if e != nil {
		j.State = "failed"
		j.Message = e.Error()
	} else {
		j.State = "complete"
		j.Message = "Ready to review or save"
	}
}
func (s *Server) getJob(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil {
		s.json(w, map[string]string{"state": "idle"})
		return
	}
	s.json(w, s.job)
}
func (s *Server) cancel(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.job != nil && s.job.State == "running" {
		s.job.cancel()
		s.job.State = "cancelling"
		s.job.Message = "Cancelling; the current result will be discarded"
	}
	s.mu.Unlock()
	s.json(w, map[string]bool{"ok": true})
}
func (s *Server) clear(w http.ResponseWriter, r *http.Request) {
	s.mu.Lock()
	if s.clearing {
		s.mu.Unlock()
		s.fail(w, errors.New("Local clearing is already in progress"))
		return
	}
	s.clearing = true
	j := s.job
	if j != nil {
		j.cancel()
		j.State = "cancelled"
	}
	s.mu.Unlock()
	s.worker.Wait()
	var cleanupErr error
	if j != nil && j.dir != "" {
		cleanupErr = os.RemoveAll(j.dir)
	}
	s.Engine.Close()
	s.mu.Lock()
	s.clearing = false
	if cleanupErr == nil {
		s.job = nil
	}
	s.mu.Unlock()
	if cleanupErr != nil {
		s.fail(w, errors.New("Temporary audio could not be completely removed; your text is still available. Close PageSpindle and retry cleanup."))
		return
	}
	s.json(w, map[string]bool{"ok": true})
}
func (s *Server) Close() {
	s.mu.Lock()
	s.clearing = true
	j := s.job
	if j != nil {
		j.cancel()
		j.State = "cancelled"
	}
	s.mu.Unlock()
	s.worker.Wait()
	s.Engine.Close()
	if j != nil && j.dir != "" {
		os.RemoveAll(j.dir)
	}
}
func (s *Server) prepare(w http.ResponseWriter, r *http.Request) {
	var d book.Document
	if e := decode(r, &d); e != nil {
		s.fail(w, e)
		return
	}
	if e := book.Validate(d); e != nil {
		s.fail(w, e)
		return
	}
	j, ctx, e := s.begin("Preparing text locally", len(d.Blocks))
	if e != nil {
		s.fail(w, e)
		return
	}
	s.json(w, map[string]int{"id": j.ID})
	go func() {
		defer s.worker.Done()
		var err error
		defer func() { s.finish(j, err) }()
		for i, b := range d.Blocks {
			if ctx.Err() != nil {
				err = ctx.Err()
				return
			}
			s.update(j, i, fmt.Sprintf("Preparing passage %d of %d", i+1, len(d.Blocks)))
			edit, e := s.Engine.Prepare(ctx, b.ID, b.Source)
			if e != nil {
				err = e
				return
			}
			d.Blocks[i].Prepared = edit.After
			d.Blocks[i].Reason = edit.Reason
			d.Blocks[i].Reviewed = false
		}
		s.mu.Lock()
		if s.job == j && ctx.Err() == nil {
			j.Document = &d
			j.Done = len(d.Blocks)
		}
		s.mu.Unlock()
	}()
}
func (s *Server) download(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.job == nil || s.job.ID != req.ID || s.job.State != "complete" {
		s.fail(w, errors.New("This result is no longer available; generate it again"))
		return
	}
	p, ok := s.job.paths[req.Name]
	if !ok {
		s.fail(w, errors.New("Unknown output file"))
		return
	}
	f, e := os.Open(p)
	if e != nil {
		s.fail(w, errors.New("The output file is unavailable"))
		return
	}
	defer f.Close()
	w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": req.Name}))
	w.Header().Set("Content-Type", mime.TypeByExtension(filepath.Ext(req.Name)))
	io.Copy(w, f)
}

type RenderRequest struct {
	Document book.Document `json:"document"`
	Speed    float64       `json:"speed"`
	Format   string        `json:"format"`
	Preview  bool          `json:"preview"`
}

func (s *Server) render(w http.ResponseWriter, r *http.Request) {
	var req RenderRequest
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	if e := book.Ready(req.Document); e != nil {
		s.fail(w, e)
		return
	}
	if req.Speed < 0.6 || req.Speed > 1.6 {
		s.fail(w, errors.New("Choose a speed from 0.6 to 1.6"))
		return
	}
	if req.Format != "wav" && req.Format != "mp3" && req.Format != "m4b" {
		s.fail(w, errors.New("Choose WAV, MP3, or M4B"))
		return
	}
	if req.Format != "wav" && !s.Engine.Status().FFmpeg {
		s.fail(w, errors.New("Install FFmpeg for MP3/M4B, or choose the built-in WAV audiobook option"))
		return
	}
	if !s.Engine.Status().Speech {
		s.fail(w, errors.New("Install the app runtime and common model pack first"))
		return
	}
	master := book.Master(req.Document)
	if len(master) > 300000 {
		s.fail(w, errors.New("Narrate this book in sections up to 300,000 UTF-8 bytes"))
		return
	}
	if req.Preview {
		r := []rune(master)
		if len(r) > 350 {
			master = string(r[:350])
		}
	}
	chapters := book.Chapters(master)
	j, ctx, e := s.begin("Generating local narration", len(chapters))
	if e != nil {
		s.fail(w, e)
		return
	}
	dir, e := os.MkdirTemp("", "pagespindle-")
	if e != nil {
		s.worker.Done()
		s.finish(j, errors.New("Cannot create private output folder; check free disk space"))
		s.fail(w, e)
		return
	}
	s.mu.Lock()
	j.dir = dir
	s.mu.Unlock()
	s.json(w, map[string]int{"id": j.ID})
	go func() { defer s.worker.Done(); err := s.makeAudio(ctx, j, req, chapters); s.finish(j, err) }()
}
func (s *Server) makeAudio(ctx context.Context, j *Job, req RenderRequest, chapters []book.Chapter) error {
	var names []string
	var total int64
	var timings []float64
	for i, ch := range chapters {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		s.update(j, i, fmt.Sprintf("Narrating chapter %d of %d", i+1, len(chapters)))
		name := fmt.Sprintf("%03d - %s.wav", i+1, book.SafeName(ch.Title))
		p := filepath.Join(j.dir, name)
		f, e := os.OpenFile(p, os.O_CREATE|os.O_EXCL|os.O_RDWR, 0600)
		if e != nil {
			return errors.New("Cannot write the chapter; check free disk space")
		}
		f.Write(make([]byte, 44))
		var n uint32
		var rate uint32
		chunks, e := book.New(ch.Text, ch.Title)
		if e != nil {
			f.Close()
			return e
		}
		for k, b := range chunks.Blocks {
			if ctx.Err() != nil {
				f.Close()
				return ctx.Err()
			}
			s.update(j, i, fmt.Sprintf("Chapter %d: passage %d of %d", i+1, k+1, len(chunks.Blocks)))
			part := filepath.Join(j.dir, "chunk.wav")
			if e = s.Engine.Speak(ctx, b.Prepared, part, req.Speed); e != nil {
				f.Close()
				return e
			}
			pcm, sr, e := book.PCMData(part)
			os.Remove(part)
			if e != nil {
				f.Close()
				return e
			}
			if rate != 0 && sr != rate {
				f.Close()
				return errors.New("The voice changed audio format unexpectedly")
			}
			rate = sr
			total += int64(len(pcm))
			if total > 768<<20 {
				f.Close()
				return errors.New("Audio exceeds 768 MiB; save this book in smaller sections")
			}
			written, err := f.Write(pcm)
			if err != nil || written != len(pcm) {
				f.Close()
				return errors.New("Chapter write failed; check free disk space")
			}
			n += uint32(len(pcm))
		}
		if e = book.WriteWAVHeader(f, rate, n); e != nil {
			f.Close()
			return e
		}
		if e = f.Close(); e != nil {
			return errors.New("Chapter could not be saved")
		}
		timings = append(timings, float64(n)/float64(rate*2))
		names = append(names, name)
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	title := book.SafeName(req.Document.Title)
	if req.Preview {
		s.mu.Lock()
		if s.job == j {
			j.Files = []string{names[0]}
			j.paths[names[0]] = filepath.Join(j.dir, names[0])
		}
		s.mu.Unlock()
		return nil
	}
	output := title + " - Audiobook.zip"
	if req.Format == "mp3" {
		for i, name := range names {
			mp3 := strings.TrimSuffix(name, ".wav") + ".mp3"
			if e := s.Engine.FFmpeg(ctx, "-nostdin", "-n", "-i", filepath.Join(j.dir, name), "-ac", "1", "-c:a", "libmp3lame", "-b:a", "96k", filepath.Join(j.dir, mp3)); e != nil {
				return e
			}
			names[i] = mp3
		}
	}
	if req.Format == "m4b" {
		var concat, meta strings.Builder
		meta.WriteString(";FFMETADATA1\ntitle=" + metaEscape(req.Document.Title) + "\nartist=" + metaEscape(req.Document.Author) + "\n")
		var ms int64
		for i, name := range names { // Names written into concat are generated internal aliases, never user paths.
			alias := fmt.Sprintf("track-%03d.wav", i)
			if e := os.Rename(filepath.Join(j.dir, name), filepath.Join(j.dir, alias)); e != nil {
				return e
			}
			concat.WriteString("file '" + alias + "'\n")
			end := ms + int64(timings[i]*1000)
			fmt.Fprintf(&meta, "[CHAPTER]\nTIMEBASE=1/1000\nSTART=%d\nEND=%d\ntitle=%s\n", ms, end, metaEscape(chapters[i].Title))
			ms = end
		}
		os.WriteFile(filepath.Join(j.dir, "chapters.txt"), []byte(concat.String()), 0600)
		os.WriteFile(filepath.Join(j.dir, "book.ffmetadata"), []byte(meta.String()), 0600)
		output = title + ".m4b"
		if e := s.Engine.FFmpeg(ctx, "-nostdin", "-n", "-f", "concat", "-safe", "1", "-i", filepath.Join(j.dir, "chapters.txt"), "-f", "ffmetadata", "-i", filepath.Join(j.dir, "book.ffmetadata"), "-map", "0:a:0", "-map_metadata", "1", "-map_chapters", "1", "-c:a", "aac", "-b:a", "96k", "-movflags", "+faststart", "-f", "mp4", filepath.Join(j.dir, output)); e != nil {
			return e
		}
	} else {
		f, e := os.OpenFile(filepath.Join(j.dir, output), os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			return e
		}
		z := zip.NewWriter(f)
		for _, name := range names {
			out, e := z.Create(name)
			if e != nil {
				z.Close()
				f.Close()
				return e
			}
			in, e := os.Open(filepath.Join(j.dir, name))
			if e != nil {
				z.Close()
				f.Close()
				return e
			}
			_, err := io.Copy(out, in)
			in.Close()
			if err != nil {
				z.Close()
				f.Close()
				return err
			}
		}
		out, e := z.Create("Listening Edition.txt")
		if e == nil {
			_, e = io.WriteString(out, book.Master(req.Document))
		}
		if e == nil {
			var p io.Writer
			p, e = z.Create("Chapters.m3u")
			if e == nil {
				_, e = io.WriteString(p, strings.Join(names, "\n")+"\n")
			}
		}
		if err := z.Close(); e == nil {
			e = err
		}
		if err := f.Close(); e == nil {
			e = err
		}
		if e != nil {
			return errors.New("Audiobook archive could not be saved; check free disk space")
		}
	}
	s.mu.Lock()
	if s.job == j && ctx.Err() == nil {
		j.Files = []string{output}
		j.paths[output] = filepath.Join(j.dir, output)
		j.Done = j.Total
	}
	s.mu.Unlock()
	return ctx.Err()
}
func metaEscape(s string) string {
	r := strings.NewReplacer("\\", "\\\\", "=", "\\=", ";", "\\;", "#", "\\#", "\n", " ", "\r", " ")
	return r.Replace(s)
}

var _ = strconv.Itoa
