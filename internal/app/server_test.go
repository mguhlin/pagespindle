package app

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/mguhlin/pagespindle/internal/book"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func call(s *Server, method, path, origin, key string, data any) *httptest.ResponseRecorder {
	var b bytes.Buffer
	if data != nil {
		json.NewEncoder(&b).Encode(data)
	}
	r := httptest.NewRequest(method, "http://"+s.Host+path, &b)
	r.Host = s.Host
	if origin != "" {
		r.Header.Set("Origin", origin)
	}
	r.Header.Set("X-PageSpindle-Key", key)
	w := httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	return w
}
func TestLocalAuthorityBoundary(t *testing.T) {
	s := New("private-test-key", "127.0.0.1:41890", t.TempDir())
	defer s.Close()
	payload := map[string]string{"text": "A fictional source.", "title": "Test"}
	for _, tc := range []struct {
		key, origin string
		status      int
	}{{"", "http://" + s.Host, 403}, {s.Key, "https://attacker.invalid", 403}, {s.Key, "", 403}, {s.Key, "http://" + s.Host, 200}} {
		w := call(s, "POST", "/api/new", tc.origin, tc.key, payload)
		if w.Code != tc.status {
			t.Fatalf("got %d want %d", w.Code, tc.status)
		}
	}
	w := call(s, "GET", "/api/new", "", s.Key, nil)
	if w.Code != 405 {
		t.Fatal("mutation allowed via GET")
	}
	r := httptest.NewRequest("GET", "http://evil.invalid/", nil)
	w = httptest.NewRecorder()
	s.Handler().ServeHTTP(w, r)
	if w.Code != 403 {
		t.Fatal("host rebinding accepted")
	}
}
func TestNarrationDoesNotRequireReview(t *testing.T) {
	s := New("key", "127.0.0.1:41890", t.TempDir())
	defer s.Close()
	d, _ := book.New("Keep this exact text.", "Test")
	w := call(s, "POST", "/api/render", "http://"+s.Host, s.Key, RenderRequest{Document: d, Speed: 1, Format: "wav"})
	if w.Code != 400 || !strings.Contains(w.Body.String(), "Install the app runtime") {
		t.Fatal("unreviewed text did not reach the normal runtime availability check")
	}
	w = call(s, "POST", "/api/download", "http://"+s.Host, s.Key, map[string]any{"id": 1, "name": "../../outside"})
	if w.Code != 400 {
		t.Fatal("unknown file accepted")
	}
}
func TestCancelCannotStartOverLiveWorker(t *testing.T) {
	s := New("key", "127.0.0.1:41890", t.TempDir())
	defer s.Close()
	j, ctx, e := s.begin("synthetic", 1)
	if e != nil {
		t.Fatal(e)
	}
	call(s, "POST", "/api/cancel", "http://"+s.Host, s.Key, map[string]bool{})
	if ctx.Err() != context.Canceled {
		t.Fatal("cancellation did not propagate")
	}
	if _, _, e = s.begin("overlap", 1); e == nil {
		t.Fatal("new job overlapped cancelling worker")
	}
	s.finish(j, ctx.Err())
	s.worker.Done()
	if j.State != "cancelled" {
		t.Fatal("cancellation reported success")
	}
}
func TestClearPreventsLateResultsAndRemovesOnlyOwnFolder(t *testing.T) {
	root := t.TempDir()
	sentinel := filepath.Join(root, "outside.txt")
	os.WriteFile(sentinel, []byte("preserve"), 0600)
	s := New("key", "127.0.0.1:41890", root)
	j, ctx, _ := s.begin("synthetic", 1)
	j.dir = filepath.Join(root, "private-output")
	os.Mkdir(j.dir, 0700)
	os.WriteFile(filepath.Join(j.dir, "audio.wav"), []byte("synthetic"), 0600)
	go func() { <-ctx.Done(); time.Sleep(10 * time.Millisecond); s.finish(j, ctx.Err()); s.worker.Done() }()
	w := call(s, "POST", "/api/clear", "http://"+s.Host, s.Key, map[string]bool{})
	if w.Code != http.StatusOK || s.job != nil {
		t.Fatal("clear incomplete")
	}
	if _, e := os.Stat(j.dir); !os.IsNotExist(e) {
		t.Fatal("temporary output remains")
	}
	if b, _ := os.ReadFile(sentinel); string(b) != "preserve" {
		t.Fatal("outside work changed")
	}
	s.update(j, 1, "late success")
	if s.job != nil {
		t.Fatal("late callback resurrected job")
	}
}

func TestPassageWorkAdvancesInsideOneChapter(t *testing.T) {
	s := New("key", "127.0.0.1:44000", t.TempDir())
	defer s.Close()
	j, _, err := s.begin("narrating", 1)
	if err != nil {
		t.Fatal(err)
	}
	defer s.worker.Done()
	s.work(j, "narrating", 250, 1000, "CPU")
	w := call(s, "GET", "/api/job", "http://"+s.Host, s.Key, nil)
	var got Job
	if json.Unmarshal(w.Body.Bytes(), &got) != nil {
		t.Fatal("invalid progress response")
	}
	if got.Done != 0 || got.Total != 1 || got.WorkDone != 250 || got.WorkTotal != 1000 || got.Device != "CPU" {
		t.Fatal("chapter-only progress lost passage work")
	}
	j.cancel()
	s.mu.Lock()
	j.State = "cancelling"
	s.mu.Unlock()
	s.work(j, "narrating", 500, 1000, "CPU")
	if j.WorkDone != 250 {
		t.Fatal("cancelled job advanced with a late result")
	}
	s.finish(j, nil)
}
