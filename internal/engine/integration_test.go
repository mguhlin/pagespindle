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
	out := filepath.Join(t.TempDir(), "preview.wav")
	if err = e.Speak(ctx, "This is PageSpindle. Spin text into listening.", out, 1); err != nil {
		t.Fatal(err)
	}
	pcm, rate, err := book.PCMData(out)
	if err != nil || rate != 22050 || len(pcm) < 22050 {
		t.Fatal("actual voice output is not valid speech-length PCM")
	}
	t.Logf("Real local model returned a complete reviewed proposal; voice generated %.2f seconds of mono PCM", float64(len(pcm))/float64(rate*2))
}
