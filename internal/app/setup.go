package app

import (
	"net/http"
)

func (s *Server) setupStatus(w http.ResponseWriter, r *http.Request) {
	if s.Setup == nil {
		http.NotFound(w, r)
		return
	}
	s.json(w, s.Setup.Snapshot())
}
func (s *Server) setupStart(w http.ResponseWriter, r *http.Request) {
	if s.Setup == nil {
		http.NotFound(w, r)
		return
	}
	if e := decode(r, &struct{}{}); e != nil {
		s.fail(w, e)
		return
	}
	s.Setup.Start()
	s.json(w, map[string]bool{"ok": true})
}
func (s *Server) setupCancel(w http.ResponseWriter, r *http.Request) {
	if s.Setup == nil {
		http.NotFound(w, r)
		return
	}
	s.Setup.Cancel()
	s.json(w, map[string]bool{"ok": true})
}
func (s *Server) windowReady(w http.ResponseWriter, r *http.Request) {
	if s.Desktop == nil {
		http.NotFound(w, r)
		return
	}
	s.readyOnce.Do(func() { close(s.NativeReady) })
	s.json(w, map[string]bool{"ok": true})
}
func (s *Server) windowReport(w http.ResponseWriter, r *http.Request) {
	if s.SelfTest == nil {
		http.NotFound(w, r)
		return
	}
	var report struct {
		OK      bool   `json:"ok"`
		Message string `json:"message"`
	}
	if e := decode(r, &report); e != nil {
		s.fail(w, e)
		return
	}
	if len(report.Message) > 1000 {
		report.Message = report.Message[:1000]
	}
	s.SelfTest(report.OK, report.Message)
	s.json(w, map[string]bool{"ok": true})
}
