package engine

import (
	"context"
	"github.com/mguhlin/pagespindle/internal/book"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	if filepath.Base(os.Args[0]) == "llama-server" || filepath.Base(os.Args[0]) == "llama-server.exe" {
		os.Exit(fakeBackendMain())
	}
	if os.Getenv("PAGESPINDLE_ENGINE_SUPERVISOR") == "1" {
		os.Exit(Supervise(os.Stdin))
	}
	os.Exit(m.Run())
}

func TestRealOfflineEngines(t *testing.T) {
	root := os.Getenv("PAGESPINDLE_RUNTIME")
	if root == "" {
		t.Skip("Set PAGESPINDLE_RUNTIME to a prepared runtime folder for real-model validation")
	}
	e := New(root)
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
	defer cancel()
	source := "The fictional library opens at 9:00 and closes at 5:00. Keep the well-known path."
	edit, err := e.Prepare(ctx, 1, source)
	if err != nil {
		t.Fatal(err)
	}
	if edit.ID != 1 || !strings.Contains(edit.After, "9:00") || !strings.Contains(edit.After, "5:00") {
		t.Fatal("model proposal changed substantive numbers")
	}
	a := e.Acceleration()
	t.Logf("Selected AI backend: %s; device: %s; fallback: %v", a.Backend, a.Device, a.Fallback)
	if os.Getenv("PAGESPINDLE_EXPECT_GPU") == "1" && a.Backend == "CPU" {
		t.Fatal("real GPU was required but CPU was selected")
	}
	out := filepath.Join(t.TempDir(), "preview-café-音.wav")
	chapters := book.Chapters("## **PageSpindle**\n\nSpin *text* into listening.")
	if len(chapters) != 1 || strings.Join(strings.Fields(chapters[0].Text), " ") != "PageSpindle Spin text into listening." {
		t.Fatal("Markdown formatting reached the speech input")
	}
	if err = e.Speak(ctx, chapters[0].Text, out, 1); err != nil {
		t.Fatal(err)
	}
	pcm, rate, err := book.PCMData(out)
	if err != nil || rate != 22050 || len(pcm) < 22050 {
		t.Fatal("actual voice output is not valid speech-length PCM")
	}
	e.mu.Lock()
	control, done, url := e.control, e.done, e.url
	e.mu.Unlock()
	control.Close()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Model supervisor did not exit after owner-pipe EOF")
	}
	if r, err := client(time.Second).Get(url + "/health"); err == nil {
		r.Body.Close()
		t.Fatal("Model listener survived owner-pipe EOF")
	}
	t.Logf("Real local model returned a complete structured proposal; voice generated %.2f seconds of mono PCM through a Unicode output path; owner-pipe EOF stopped the model listener", float64(len(pcm))/float64(rate*2))
}
