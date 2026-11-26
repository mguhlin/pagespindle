package engine

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

type supervisorSpec struct {
	Root, Models string
	Port         int
	Key          string
	ForceCPU     bool
}

func cleanEnvironment() []string {
	allowed := map[string]bool{"PATH": true, "SYSTEMROOT": true, "WINDIR": true, "TEMP": true, "TMP": true, "TMPDIR": true, "LANG": true, "LC_ALL": true}
	var env []string
	for _, v := range os.Environ() {
		k, _, _ := strings.Cut(v, "=")
		if allowed[strings.ToUpper(k)] {
			env = append(env, v)
		}
	}
	return env
}
func backendArguments(e *Engine, s supervisorSpec, b backendChoice) []string {
	args := []string{"-m", e.model(), "--host", "127.0.0.1", "--port", strconv.Itoa(s.Port), "--api-key", s.Key, "--ctx-size", "4096", "--parallel", "1", "--threads", "4", "--no-webui", "--no-ui-mcp-proxy", "--log-disable"}
	if b.backend == "CPU" {
		return append(args, "--device", "none", "--gpu-layers", "0", "--no-op-offload", "--no-kv-offload")
	}
	return append(args, "--device", b.device, "--gpu-layers", "auto")
}

// The supervisor owns every probe and model child. Closing the parent's private
// pipe cancels detection, startup, inference and any pending CPU fallback.
func Supervise(in io.Reader) int {
	var spec supervisorSpec
	if json.NewDecoder(io.LimitReader(in, 8192)).Decode(&spec) != nil || !filepath.IsAbs(spec.Root) || spec.Port < 1 || spec.Port > 65535 || len(spec.Key) != 64 {
		return 2
	}
	if spec.Models != "" && !filepath.IsAbs(spec.Models) {
		return 2
	}
	e := New(spec.Root)
	e.ModelRoot = spec.Models
	e.ForceCPU = spec.ForceCPU
	if !exists(e.path("llama", "llama-server")) || !exists(e.model()) {
		return 3
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() { io.Copy(io.Discard, in); cancel() }()
	choice := e.preferredBackend(ctx)
	choices := []backendChoice{choice}
	if choice.backend != "CPU" {
		choices = append(choices, backendChoice{role: "llama", backend: "CPU", name: "CPU"})
	}
	reporter := json.NewEncoder(os.Stdout)
	for i, b := range choices {
		if ctx.Err() != nil {
			return 0
		}
		report := Acceleration{Backend: b.backend, Device: b.name, Stage: "loading", Fallback: i > 0}
		reporter.Encode(report)
		program := e.path(b.role, "llama-server")
		c := child(ctx, program, backendArguments(e, spec, b)...)
		c.Stdout = io.Discard
		c.Stderr = io.Discard
		if err := c.Start(); err != nil {
			continue
		}
		done := make(chan error, 1)
		go func() { done <- c.Wait() }()
		limit := 3 * time.Minute
		if b.backend != "CPU" {
			limit = 90 * time.Second
		}
		deadline := time.NewTimer(limit)
		tick := time.NewTicker(200 * time.Millisecond)
		ready := false
		failed := false
		for !ready && !failed {
			select {
			case <-ctx.Done():
				killProcess(c)
				<-done
				deadline.Stop()
				tick.Stop()
				return 0
			case <-deadline.C:
				killProcess(c)
				<-done
				failed = true
			case <-done:
				failed = true
			case <-tick.C:
				r, err := client(time.Second).Get("http://127.0.0.1:" + strconv.Itoa(spec.Port) + "/health")
				if err == nil {
					r.Body.Close()
					ready = r.StatusCode == 200
				}
			}
		}
		deadline.Stop()
		tick.Stop()
		if failed {
			continue
		}
		report.Stage = "ready"
		reporter.Encode(report)
		select {
		case <-ctx.Done():
			killProcess(c)
			<-done
			return 0
		case err := <-done:
			if err == nil {
				return 0
			}
			if errors.Is(ctx.Err(), context.Canceled) {
				return 0
			}
		}
	}
	return 5
}
