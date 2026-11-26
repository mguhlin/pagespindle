package engine

import (
	"context"
	"regexp"
	"runtime"
	"strings"
	"time"
)

type Acceleration struct {
	Backend  string `json:"backend"`
	Device   string `json:"device"`
	Stage    string `json:"stage"`
	Fallback bool   `json:"fallback"`
}
type backendChoice struct{ role, backend, device, name string }

var deviceLine = regexp.MustCompile(`^\s*((?:Vulkan|MTL)[0-9]+):\s+(.+?)\s+\([0-9]+ MiB,`)

func chooseDevice(output string, prefix string) (string, string) {
	best := -1
	id, name := "", ""
	for _, line := range strings.Split(output, "\n") {
		m := deviceLine.FindStringSubmatch(line)
		if len(m) != 3 || !strings.HasPrefix(m[1], prefix) {
			continue
		}
		lower := strings.ToLower(m[2])
		software := false
		for _, s := range []string{"llvmpipe", "lavapipe", "swiftshader", "software", "microsoft basic", "cpu"} {
			if strings.Contains(lower, s) {
				software = true
			}
		}
		if software {
			continue
		}
		score := 1
		for _, s := range []string{"nvidia", "geforce", "quadro", "tesla", "radeon rx", "radeon pro", "radeon vii", "arc"} {
			if strings.Contains(lower, s) {
				score = 100
				if s == "arc" {
					score = 70
				}
				break
			}
		}
		if score > best {
			best = score
			id = m[1]
			name = m[2]
		}
	}
	return id, name
}

type cappedOutput struct{ data []byte }

func (b *cappedOutput) Write(p []byte) (int, error) {
	n := len(p)
	left := 65536 - len(b.data)
	if left > 0 {
		if len(p) > left {
			p = p[:left]
		}
		b.data = append(b.data, p...)
	}
	return n, nil
}
func (e *Engine) preferredBackend(ctx context.Context) backendChoice {
	cpu := backendChoice{role: "llama", backend: "CPU", name: "CPU"}
	if e.ForceCPU || !e.gpuRuntimeAvailable() {
		return cpu
	}
	role, backend := "llama-gpu", "Vulkan"
	if runtime.GOOS == "darwin" {
		backend = "Metal"
		if runtime.GOARCH == "arm64" {
			role = "llama"
		}
		if check := e.path("gpu-support", "pagespindle-metal-check"); exists(check) {
			pctx, cancel := context.WithTimeout(ctx, 5*time.Second)
			cmd := child(pctx, check)
			err := cmd.Run()
			cancel()
			if err != nil {
				return cpu
			}
		}
	}
	program := e.path(role, "llama-server")
	if !exists(program) {
		return cpu
	}
	pctx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	cmd := child(pctx, program, "--list-devices")
	var output cappedOutput
	cmd.Stdout = &output
	if cmd.Run() != nil {
		return cpu
	}
	prefix := backend
	if backend == "Metal" {
		prefix = "MTL"
	}
	id, name := chooseDevice(string(output.data), prefix)
	if id == "" {
		return cpu
	}
	return backendChoice{role: role, backend: backend, device: id, name: name}
}
func (e *Engine) Acceleration() Acceleration {
	e.accelMu.RLock()
	defer e.accelMu.RUnlock()
	a := e.acceleration
	if a.Backend == "" {
		a = Acceleration{Backend: "Automatic", Device: "GPU checked when AI starts", Stage: "idle"}
	}
	return a
}
func (e *Engine) AccelerationLabel() string {
	a := e.Acceleration()
	if a.Backend == "CPU" {
		if a.Fallback {
			return "CPU (GPU unavailable or unable to run this model)"
		}
		return "CPU"
	}
	if a.Backend == "Automatic" {
		return "Automatic GPU / CPU"
	}
	return a.Backend + " — " + a.Device
}
func (e *Engine) setAcceleration(epoch uint64, a Acceleration) {
	if len(a.Device) > 200 || len(a.Stage) > 32 {
		return
	}
	switch a.Backend {
	case "CPU", "Metal", "Vulkan":
	default:
		return
	}
	e.accelMu.Lock()
	defer e.accelMu.Unlock()
	if e.accelEpoch.Load() == epoch {
		e.acceleration = a
	}
}
