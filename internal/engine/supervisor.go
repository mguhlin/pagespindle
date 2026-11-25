package engine

import (
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type supervisorSpec struct {
	Root   string
	Models string
	Port   int
	Key    string
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

// Supervise owns exactly one bundled model process. EOF on the parent's private
// pipe (including abrupt parent death) terminates it. There is one Wait owner.
func Supervise(in io.Reader) int {
	var spec supervisorSpec
	if json.NewDecoder(io.LimitReader(in, 8192)).Decode(&spec) != nil || !filepath.IsAbs(spec.Root) || spec.Port < 1 || spec.Port > 65535 || len(spec.Key) != 64 {
		return 2
	}
	e := New(spec.Root)
	if spec.Models != "" && !filepath.IsAbs(spec.Models) {
		return 2
	}
	e.ModelRoot = spec.Models
	program := e.path("llama", "llama-server")
	if !exists(program) || !exists(e.model()) {
		return 3
	}
	c := exec.Command(program, "-m", e.model(), "--host", "127.0.0.1", "--port", strconv.Itoa(spec.Port), "--api-key", spec.Key, "--ctx-size", "4096", "--parallel", "1", "--threads", "4", "--no-webui", "--no-ui-mcp-proxy", "--log-disable")
	c.Env = cleanEnvironment()
	c.Stdout = io.Discard
	c.Stderr = io.Discard
	if c.Start() != nil {
		return 4
	}
	done := make(chan error, 1)
	go func() { done <- c.Wait() }()
	closed := make(chan struct{})
	go func() { io.Copy(io.Discard, in); close(closed) }()
	select {
	case err := <-done:
		if err != nil {
			return 5
		}
		return 0
	case <-closed:
		c.Process.Kill()
		<-done
		return 0
	}
}
