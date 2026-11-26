package engine

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"
)

func TestDeviceSelection(t *testing.T) {
	id, name := chooseDevice("Available devices:\n Vulkan0: llvmpipe (123 MiB, 123 MiB free)\n Vulkan1: Intel(R) UHD Graphics 630 (10000 MiB, 5000 MiB free)\n Vulkan2: NVIDIA GeForce RTX 2060 (6000 MiB, 5000 MiB free)", "Vulkan")
	if id != "Vulkan2" || !strings.Contains(name, "RTX") {
		t.Fatal("discrete hardware GPU was not selected")
	}
	if id, _ = chooseDevice("Vulkan0: SwiftShader (100 MiB, 100 MiB free)", "Vulkan"); id != "" {
		t.Fatal("software renderer counted as GPU")
	}
	if id, _ = chooseDevice("malformed GPU data", "Metal"); id != "" {
		t.Fatal("malformed probe accepted")
	}
}
func TestMetalDeviceSelection(t *testing.T) {
	id, _ := chooseDevice("  MTL0: Apple M1 (6000 MiB, 5000 MiB free)", "MTL")
	if id != "MTL0" {
		t.Fatal("Metal device not recognized")
	}
}
func TestCPUArgumentsDisableOffload(t *testing.T) {
	args := strings.Join(backendArguments(New(t.TempDir()), supervisorSpec{Port: 12345, Key: strings.Repeat("a", 64)}, backendChoice{backend: "CPU"}), " ")
	for _, want := range []string{"--host 127.0.0.1", "--device none", "--gpu-layers 0", "--no-op-offload", "--no-kv-offload"} {
		if !strings.Contains(args, want) {
			t.Fatal("missing CPU confinement option", want)
		}
	}
}
func fakeGPUFixture(t *testing.T, mode string) *Engine {
	t.Helper()
	root := t.TempDir()
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	b, err := os.ReadFile(exe)
	if err != nil {
		t.Fatal(err)
	}
	name := "llama-server"
	if runtime.GOOS == "windows" {
		name += ".exe"
	}
	for _, role := range []string{"llama", "llama-gpu"} {
		p := filepath.Join(root, "runtime", runtime.GOOS+"-"+runtime.GOARCH, role, name)
		os.MkdirAll(filepath.Dir(p), 0700)
		if err = os.WriteFile(p, b, 0700); err != nil {
			t.Fatal(err)
		}
	}
	if runtime.GOOS == "windows" {
		os.WriteFile(filepath.Join(root, "runtime", runtime.GOOS+"-"+runtime.GOARCH, "llama-gpu", "vulkan-1.dll"), []byte("fixture loader marker; fake backend is Go"), 0600)
	}
	os.Mkdir(filepath.Join(root, "models"), 0700)
	os.WriteFile(filepath.Join(root, "models", "Qwen3-1.7B-Q8_0.gguf"), []byte(mode), 0600)
	return New(root)
}
func TestGPUStartupFailureFallsBackToCPU(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("fixture uses Vulkan role; real Metal follows Mac platform checks")
	}
	e := fakeGPUFixture(t, "startup-failure")
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.start(ctx); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(time.Second)
	for e.Acceleration().Stage != "ready" && time.Now().Before(deadline) {
		time.Sleep(time.Millisecond * 10)
	}
	a := e.Acceleration()
	if a.Backend != "CPU" || !a.Fallback {
		t.Fatalf("GPU failure did not become a CPU fallback: %+v", a)
	}
}
func TestGPUInferenceFailureRetriesOnceOnCPU(t *testing.T) {
	if runtime.GOOS == "darwin" {
		t.Skip("fixture uses Vulkan role")
	}
	e := fakeGPUFixture(t, "inference-failure")
	defer e.Close()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	edit, err := e.Prepare(ctx, 1, "Fictional protected source.")
	if err != nil {
		t.Fatal(err)
	}
	if edit.After != "Fictional protected source." || e.Acceleration().Backend != "CPU" {
		t.Fatal("fallback lost source or continued on failing GPU")
	}
}

// Invoked only when the test binary was copied to its private fake engine path.
func fakeBackendMain() int {
	gpu := strings.Contains(filepath.Dir(os.Args[0]), "llama-gpu")
	for _, a := range os.Args[1:] {
		if a == "--list-devices" {
			if gpu {
				os.Stdout.WriteString("Available devices:\n Vulkan0: NVIDIA Synthetic Test Device (6000 MiB, 5000 MiB free)\n")
			}
			return 0
		}
	}
	values := map[string]string{}
	for i := 1; i+1 < len(os.Args); i++ {
		values[os.Args[i]] = os.Args[i+1]
	}
	b, _ := os.ReadFile(values["-m"])
	if gpu && string(b) == "startup-failure" {
		return 7
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { w.Write([]byte(`{"status":"ok"}`)) })
	mux.HandleFunc("/v1/chat/completions", func(w http.ResponseWriter, r *http.Request) {
		if gpu {
			w.WriteHeader(500)
			return
		}
		json.NewEncoder(w).Encode(map[string]any{"choices": []any{map[string]any{"message": map[string]string{"content": `{"id":1,"after":"Fictional protected source.","reason":"Original retained"}`}, "finish_reason": "stop"}}})
	})
	if http.ListenAndServe("127.0.0.1:"+values["--port"], mux) != nil {
		return 8
	}
	return 0
}
