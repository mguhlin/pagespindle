package app

import (
	"errors"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

// The picker is supplied by the native host. Renderer strings can suggest a
// filename, but cannot select a filesystem path or bypass the OS save dialog.
type DesktopHooks struct {
	ChooseSave func(string) (string, error)
	OpenHelp   func(string) error
}

func (s *Server) openHelp(w http.ResponseWriter, r *http.Request) {
	if s.Desktop == nil || s.Desktop.OpenHelp == nil {
		http.NotFound(w, r)
		return
	}
	var req struct {
		ID string `json:"id"`
	}
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	urls := map[string]string{"vcc": "https://mguhlin.github.io/vcc/guides/skills.html#audiobook-maker", "source": "https://github.com/mguhlin/pagespindle"}
	url, ok := urls[req.ID]
	if !ok {
		s.fail(w, errors.New("Unknown help destination"))
		return
	}
	if e := s.Desktop.OpenHelp(url); e != nil {
		s.fail(w, errors.New("The help page could not open"))
		return
	}
	s.json(w, map[string]bool{"ok": true})
}

func (s *Server) DesktopDirty() bool { return s.desktopDirty.Load() }
func (s *Server) desktopState(w http.ResponseWriter, r *http.Request) {
	if s.Desktop == nil {
		http.NotFound(w, r)
		return
	}
	var req struct {
		Dirty bool `json:"dirty"`
	}
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	s.desktopDirty.Store(req.Dirty)
	s.json(w, map[string]bool{"ok": true})
}
func suggestedName(name string) string {
	name = filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	if name == "." || name == "/" || strings.ContainsAny(name, "\x00\r\n") || len(name) > 240 {
		return "PageSpindle-output.txt"
	}
	return name
}
func saveNew(path string, source io.Reader, max int64) error {
	if !filepath.IsAbs(path) {
		return errors.New("The native dialog did not return an absolute destination")
	}
	f, e := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600)
	if e != nil {
		if os.IsExist(e) {
			return errors.New("A file with that name already exists. Choose a new filename; the existing file was preserved")
		}
		return errors.New("Cannot create the selected file; choose a writable folder")
	}
	success := false
	defer func() {
		f.Close()
		if !success {
			os.Remove(path)
		}
	}()
	n, e := io.Copy(f, io.LimitReader(source, max+1))
	if e != nil || n > max {
		return errors.New("The output could not be saved completely; the new partial file was removed")
	}
	if e = f.Sync(); e != nil {
		return errors.New("The output could not be flushed to disk")
	}
	if e = f.Close(); e != nil {
		return errors.New("The output could not be closed")
	}
	success = true
	return nil
}
func (s *Server) saveText(w http.ResponseWriter, r *http.Request) {
	if s.Desktop == nil {
		http.NotFound(w, r)
		return
	}
	epoch := s.desktopEpoch.Load()
	name := suggestedName(r.URL.Query().Get("name"))
	data, e := io.ReadAll(io.LimitReader(r.Body, 16<<20+1))
	if e != nil || len(data) > 16<<20 {
		s.fail(w, errors.New("The text/project export exceeds 16 MiB"))
		return
	}
	path, e := s.Desktop.ChooseSave(name)
	if e != nil {
		s.fail(w, errors.New("The native save dialog could not open"))
		return
	}
	if path == "" {
		s.json(w, map[string]bool{"cancelled": true})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.desktopEpoch.Load() != epoch {
		s.fail(w, errors.New("The edition changed while the save dialog was open; save the current edition again"))
		return
	}
	if e = saveNew(path, strings.NewReader(string(data)), 16<<20); e != nil {
		s.fail(w, e)
		return
	}
	s.json(w, map[string]any{"saved": true, "name": filepath.Base(path)})
}
func (s *Server) saveJob(w http.ResponseWriter, r *http.Request) {
	if s.Desktop == nil {
		http.NotFound(w, r)
		return
	}
	var req struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}
	if e := decode(r, &req); e != nil {
		s.fail(w, e)
		return
	}
	epoch := s.desktopEpoch.Load()
	s.mu.Lock()
	valid := s.job != nil && s.job.ID == req.ID && s.job.State == "complete"
	if valid {
		_, valid = s.job.paths[req.Name]
	}
	s.mu.Unlock()
	if !valid {
		s.fail(w, errors.New("This audiobook output is no longer available"))
		return
	}
	path, e := s.Desktop.ChooseSave(suggestedName(req.Name))
	if e != nil {
		s.fail(w, errors.New("The native save dialog could not open"))
		return
	}
	if path == "" {
		s.json(w, map[string]bool{"cancelled": true})
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.desktopEpoch.Load() != epoch || s.job == nil || s.job.ID != req.ID || s.job.State != "complete" {
		s.fail(w, errors.New("The audiobook changed while the save dialog was open; generate or select it again"))
		return
	}
	source, ok := s.job.paths[req.Name]
	if !ok {
		s.fail(w, errors.New("Unknown audiobook output"))
		return
	}
	f, e := os.Open(source)
	if e != nil {
		s.fail(w, errors.New("The audiobook output is unavailable"))
		return
	}
	defer f.Close()
	if e = saveNew(path, f, 800<<20); e != nil {
		s.fail(w, e)
		return
	}
	s.json(w, map[string]any{"saved": true, "name": filepath.Base(path)})
}
