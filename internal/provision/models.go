package provision

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Entry struct {
	Path  string `json:"path"`
	SHA   string `json:"sha256"`
	Bytes int64  `json:"bytes"`
}
type receipt struct {
	Pack  string  `json:"pack"`
	Files []Entry `json:"files"`
}
type State struct {
	Phase   string `json:"phase"`
	Message string `json:"message"`
	Bytes   int64  `json:"bytes"`
	Total   int64  `json:"total"`
}
type Manager struct {
	Directory string
	mu        sync.Mutex
	state     State
	cancel    context.CancelFunc
	done      chan struct{}
}

func New(directory string) *Manager {
	return &Manager{Directory: directory, state: State{Phase: "idle", Total: PackBytes, Message: "Preparing first-launch model setup"}}
}
func (m *Manager) Snapshot() State { m.mu.Lock(); defer m.mu.Unlock(); return m.state }
func (m *Manager) update(phase, message string, n int64) {
	m.mu.Lock()
	m.state = State{Phase: phase, Message: message, Bytes: n, Total: PackBytes}
	m.mu.Unlock()
}
func digestFile(p string) (string, error) {
	f, e := os.Open(p)
	if e != nil {
		return "", e
	}
	defer f.Close()
	h := sha256.New()
	_, e = io.Copy(h, f)
	return hex.EncodeToString(h.Sum(nil)), e
}

type cancelReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r cancelReader) Read(b []byte) (int, error) {
	if e := r.ctx.Err(); e != nil {
		return 0, e
	}
	return r.reader.Read(b)
}
func safeRelative(s string) bool {
	return s != "" && path.Clean(s) == s && !strings.ContainsAny(s, "\\\x00:") && !strings.HasPrefix(s, "/") && s != ".." && !strings.HasPrefix(s, "../")
}
func Ready(directory string) bool {
	b, e := os.ReadFile(filepath.Join(directory, ".pagespindle-models.json"))
	if e != nil || len(b) > 1<<20 {
		return false
	}
	var r receipt
	if json.Unmarshal(b, &r) != nil || r.Pack != PackSHA || len(r.Files) == 0 || len(r.Files) > 5000 {
		return false
	}
	for _, entry := range r.Files {
		if !safeRelative(entry.Path) {
			return false
		}
		p := filepath.Join(directory, filepath.FromSlash(entry.Path))
		info, e := os.Lstat(p)
		if e != nil || !info.Mode().IsRegular() || info.Size() != entry.Bytes {
			return false
		}
		sum, e := digestFile(p)
		if e != nil || sum != entry.SHA {
			return false
		}
	}
	return true
}
func (m *Manager) Start() {
	m.mu.Lock()
	if m.cancel != nil || m.state.Phase == "complete" {
		m.mu.Unlock()
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	m.cancel = cancel
	m.done = make(chan struct{})
	done := m.done
	m.mu.Unlock()
	go func() {
		defer func() { m.mu.Lock(); m.cancel = nil; close(done); m.mu.Unlock() }()
		e := m.install(ctx)
		if e != nil {
			if ctx.Err() != nil {
				m.update("cancelled", "Setup cancelled. Existing model files were preserved. Retry when ready.", 0)
			} else {
				m.update("failed", e.Error(), 0)
			}
		} else {
			m.update("complete", "Models are installed. PageSpindle can now work offline.", PackBytes)
		}
	}()
}
func (m *Manager) Cancel() {
	m.mu.Lock()
	cancel := m.cancel
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
}
func (m *Manager) Close() {
	m.mu.Lock()
	cancel, done := m.cancel, m.done
	m.mu.Unlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}
func (m *Manager) install(ctx context.Context) error {
	parent := filepath.Dir(m.Directory)
	if e := os.MkdirAll(parent, 0700); e != nil {
		return errors.New("Cannot create PageSpindle's model folder. Check disk space and folder permissions")
	}
	stage, e := os.MkdirTemp(parent, ".pagespindle-setup-")
	if e != nil {
		return e
	}
	defer os.RemoveAll(stage)
	m.update("downloading", "Downloading local models once (about 1.9 GB). Your text is never uploaded.", 0)
	client := &http.Client{Timeout: 45 * time.Minute, Transport: &http.Transport{Proxy: nil}, CheckRedirect: func(r *http.Request, via []*http.Request) error {
		if r.URL.Scheme != "https" || len(via) > 8 {
			return errors.New("Invalid model download redirect")
		}
		return nil
	}}
	req, _ := http.NewRequestWithContext(ctx, "GET", PackURL, nil)
	response, e := client.Do(req)
	if e != nil {
		return errors.New("Model download could not start. Check the internet connection, then retry")
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return errors.New("The model download server did not return the model pack. Retry later")
	}
	archive := filepath.Join(stage, "models.zip")
	out, e := os.OpenFile(archive, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		return e
	}
	h := sha256.New()
	buf := make([]byte, 1024*1024)
	var total int64
	for {
		if ctx.Err() != nil {
			out.Close()
			return ctx.Err()
		}
		n, err := response.Body.Read(buf)
		if n > 0 {
			total += int64(n)
			if total > PackBytes {
				out.Close()
				return errors.New("Model download exceeds its published size")
			}
			if _, e = out.Write(buf[:n]); e != nil {
				out.Close()
				return errors.New("The download could not be saved. Check free disk space and retry")
			}
			h.Write(buf[:n])
			m.update("downloading", "Downloading local models once (about 1.9 GB). Your text is never uploaded.", total)
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			out.Close()
			return errors.New("Model download was interrupted. Retry when the connection is available")
		}
	}
	if e = out.Close(); e != nil {
		return e
	}
	if total != PackBytes || hex.EncodeToString(h.Sum(nil)) != PackSHA {
		return errors.New("The model pack failed its checksum check. No model files were installed; retry the download")
	}
	m.update("installing", "Verifying and installing local models…", total)
	if e = extractModels(ctx, archive, stage, PackSHA); e != nil {
		return e
	}
	if ctx.Err() != nil {
		return ctx.Err()
	}
	candidate := filepath.Join(stage, "models")
	var backup string
	if _, e = os.Lstat(m.Directory); e == nil {
		backup, e = os.MkdirTemp(parent, "models-preserved-")
		if e != nil {
			return e
		}
		os.Remove(backup)
		if e = os.Rename(m.Directory, backup); e != nil {
			return errors.New("Close other PageSpindle windows, then retry model setup")
		}
	}
	if e = os.Rename(candidate, m.Directory); e != nil {
		if backup != "" {
			os.Rename(backup, m.Directory)
		}
		return errors.New("Models could not be committed to the selected folder; existing models were preserved")
	}
	return nil
}
func extractModels(ctx context.Context, archive, stage, packSHA string) error {
	z, e := zip.OpenReader(archive)
	if e != nil {
		return e
	}
	defer z.Close()
	if len(z.File) > 10000 {
		return errors.New("Too many model pack entries")
	}
	var inventory struct {
		Files []Entry `json:"files"`
	}
	found := false
	for _, f := range z.File {
		if f.Name == "PageSpindle/MODELS_INVENTORY.json" {
			r, e := f.Open()
			if e != nil {
				return e
			}
			b, e := io.ReadAll(io.LimitReader(r, 1<<20+1))
			r.Close()
			if e != nil || len(b) > 1<<20 || json.Unmarshal(b, &inventory) != nil {
				return errors.New("Invalid model inventory")
			}
			found = true
		}
	}
	if !found {
		return errors.New("The model inventory is missing")
	}
	wanted := map[string]Entry{}
	seenFold := map[string]bool{}
	var expanded int64
	var entries []Entry
	for _, entry := range inventory.Files {
		if !strings.HasPrefix(entry.Path, "models/") {
			continue
		}
		relative := strings.TrimPrefix(entry.Path, "models/")
		if !safeRelative(relative) || entry.Bytes < 0 || entry.Bytes > 2<<30 || len(entry.SHA) != 64 || seenFold[strings.ToLower(relative)] {
			return errors.New("Unsafe model inventory")
		}
		seenFold[strings.ToLower(relative)] = true
		expanded += entry.Bytes
		if expanded > 3<<30 {
			return errors.New("Model pack expansion limit")
		}
		wanted["PageSpindle/"+entry.Path] = entry
		entries = append(entries, Entry{Path: relative, SHA: entry.SHA, Bytes: entry.Bytes})
	}
	if len(wanted) == 0 {
		return errors.New("No model data in the model pack")
	}
	seen := map[string]bool{}
	for _, f := range z.File {
		if ctx.Err() != nil {
			return ctx.Err()
		}
		entry, ok := wanted[f.Name]
		if !ok {
			continue
		}
		if seen[f.Name] || !f.Mode().IsRegular() || int64(f.UncompressedSize64) != entry.Bytes {
			return errors.New("Invalid model archive entry")
		}
		seen[f.Name] = true
		relative := strings.TrimPrefix(entry.Path, "models/")
		dest := filepath.Join(stage, "models", filepath.FromSlash(relative))
		if e = os.MkdirAll(filepath.Dir(dest), 0700); e != nil {
			return e
		}
		in, e := f.Open()
		if e != nil {
			return e
		}
		out, e := os.OpenFile(dest, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0600)
		if e != nil {
			in.Close()
			return e
		}
		h := sha256.New()
		n, copyErr := io.Copy(io.MultiWriter(out, h), io.LimitReader(cancelReader{ctx, in}, entry.Bytes+1))
		in.Close()
		closeErr := out.Close()
		if ctx.Err() != nil {
			return ctx.Err()
		}
		if copyErr != nil || closeErr != nil || n != entry.Bytes || hex.EncodeToString(h.Sum(nil)) != entry.SHA {
			return errors.New("A model component failed verification")
		}
	}
	if len(seen) != len(wanted) {
		return errors.New("A model component is missing")
	}
	b, _ := json.Marshal(receipt{Pack: packSHA, Files: entries})
	return os.WriteFile(filepath.Join(stage, "models", ".pagespindle-models.json"), b, 0600)
}
