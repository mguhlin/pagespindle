package provision

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func modelFixture(t *testing.T, name string, corrupt bool) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "pack.zip")
	f, e := os.Create(p)
	if e != nil {
		t.Fatal(e)
	}
	z := zip.NewWriter(f)
	data := []byte("synthetic model, not executable")
	h := sha256.Sum256(data)
	sum := hex.EncodeToString(h[:])
	if corrupt {
		sum = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	b, _ := json.Marshal(map[string]any{"files": []Entry{{Path: "models/" + name, SHA: sum, Bytes: int64(len(data))}}})
	w, _ := z.Create("PageSpindle/MODELS_INVENTORY.json")
	w.Write(b)
	w, _ = z.Create("PageSpindle/models/" + name)
	w.Write(data)
	z.Close()
	f.Close()
	return p
}
func TestExtractVerifiedModelsAndTamper(t *testing.T) {
	stage := t.TempDir()
	if e := extractModels(context.Background(), modelFixture(t, "voice/data.bin", false), stage, PackSHA); e != nil {
		t.Fatal(e)
	}
	dir := filepath.Join(stage, "models")
	if !Ready(dir) {
		t.Fatal("verified receipt was rejected")
	}
	os.WriteFile(filepath.Join(dir, "voice/data.bin"), []byte("changed"), 0600)
	if Ready(dir) {
		t.Fatal("tampered model accepted")
	}
}
func TestExtractRejectsTraversalAndCorruption(t *testing.T) {
	for _, name := range []string{"../../outside", "/outside", "voice/../outside", "voice\\outside", "C:/outside", "voice/file:stream"} {
		t.Run(name, func(t *testing.T) {
			stage := t.TempDir()
			sentinel := filepath.Join(stage, "outside")
			os.WriteFile(sentinel, []byte("preserve"), 0600)
			if e := extractModels(context.Background(), modelFixture(t, name, false), stage, PackSHA); e == nil {
				t.Fatal("unsafe path accepted")
			}
			b, _ := os.ReadFile(sentinel)
			if string(b) != "preserve" {
				t.Fatal("outside file changed")
			}
		})
	}
	if e := extractModels(context.Background(), modelFixture(t, "voice/data.bin", true), t.TempDir(), PackSHA); e == nil {
		t.Fatal("corrupt data accepted")
	}
}
func TestExtractCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if e := extractModels(ctx, modelFixture(t, "voice/data.bin", false), t.TempDir(), PackSHA); e != context.Canceled {
		t.Fatalf("cancel did not reach extraction: %v", e)
	}
}
func TestActualReleasedPack(t *testing.T) {
	p := os.Getenv("PAGESPINDLE_MODEL_ARCHIVE")
	if p == "" {
		t.Skip("set PAGESPINDLE_MODEL_ARCHIVE for real released pack verification")
	}
	sum, e := digestFile(p)
	if e != nil || sum != PackSHA {
		t.Fatalf("release archive checksum: %v", e)
	}
	stage := t.TempDir()
	if e = extractModels(context.Background(), p, stage, PackSHA); e != nil {
		t.Fatal(e)
	}
	if !Ready(filepath.Join(stage, "models")) {
		t.Fatal("real installed pack failed verification")
	}
}

func TestCancelSetupPreservesExistingModels(t *testing.T) {
	parent := t.TempDir()
	dir := filepath.Join(parent, "models")
	os.Mkdir(dir, 0700)
	sentinel := filepath.Join(dir, "original.txt")
	os.WriteFile(sentinel, []byte("preserve"), 0600)
	m := New(dir)
	m.Start()
	m.Cancel()
	m.Close()
	if m.Snapshot().Phase != "cancelled" {
		t.Fatal("cancel state missing")
	}
	b, _ := os.ReadFile(sentinel)
	if string(b) != "preserve" {
		t.Fatal("existing model data changed")
	}
	entries, _ := os.ReadDir(parent)
	if len(entries) != 1 {
		t.Fatal("cancel left model staging behind")
	}
}
