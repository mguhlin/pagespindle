package app

import (
	"bytes"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func nativeText(s *Server, text string) *httptest.ResponseRecorder {
	r := httptest.NewRequest("POST", "http://"+s.Host+"/api/save-text?name=..%2Fexport.txt", strings.NewReader(text))
	r.Host = s.Host
	r.Header.Set("Origin", "http://"+s.Host)
	r.Header.Set("X-PageSpindle-Key", s.Key)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func TestNativeSaveRequiresPickerAndPreservesExistingFile(t *testing.T) {
	root := t.TempDir()
	s := New("key", "127.0.0.1:44000", root)
	defer s.Close()
	path := filepath.Join(root, "chosen.txt")
	os.WriteFile(path, []byte("preserve original"), 0600)
	calls := 0
	s.Desktop = &DesktopHooks{ChooseSave: func(name string) (string, error) {
		calls++
		if name != "export.txt" {
			t.Error("untrusted path reached picker default")
		}
		return path, nil
	}}
	w := nativeText(s, "new content")
	if w.Code != 400 || calls != 1 {
		t.Fatal("existing target was not rejected after native picker")
	}
	b, _ := os.ReadFile(path)
	if !bytes.Equal(b, []byte("preserve original")) {
		t.Fatal("existing file changed")
	}
	s.Desktop.ChooseSave = func(string) (string, error) { return "", nil }
	if w = nativeText(s, "cancelled"); w.Code != 200 || !strings.Contains(w.Body.String(), "cancelled") {
		t.Fatal("picker cancellation not reported")
	}
}
func TestClearInvalidatesLateNativeSave(t *testing.T) {
	root := t.TempDir()
	s := New("key", "127.0.0.1:44000", root)
	defer s.Close()
	entered, release, done := make(chan struct{}), make(chan struct{}), make(chan *httptest.ResponseRecorder)
	path := filepath.Join(root, "late.txt")
	s.Desktop = &DesktopHooks{ChooseSave: func(string) (string, error) { close(entered); <-release; return path, nil }}
	go func() { done <- nativeText(s, "must not be saved") }()
	<-entered
	w := call(s, "POST", "/api/clear", "http://"+s.Host, s.Key, map[string]bool{})
	if w.Code != 200 {
		t.Fatal("clear failed")
	}
	close(release)
	w = <-done
	if w.Code != 400 {
		t.Fatal("late native selection was accepted")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatal("late output was written")
	}
}
func TestNativeSaveRejectsForgedJobPath(t *testing.T) {
	s := New("key", "127.0.0.1:44000", t.TempDir())
	defer s.Close()
	called := false
	s.Desktop = &DesktopHooks{ChooseSave: func(string) (string, error) { called = true; return "", nil }}
	w := call(s, "POST", "/api/save-job", "http://"+s.Host, s.Key, map[string]any{"id": 1, "name": "../../outside", "path": "/etc/passwd"})
	if w.Code != 400 || called {
		t.Fatal("renderer forged native path authority")
	}
}
