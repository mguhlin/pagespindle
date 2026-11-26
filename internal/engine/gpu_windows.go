//go:build windows

package engine

import (
	"os"
	"path/filepath"
)

func (e *Engine) gpuRuntimeAvailable() bool {
	// The Vulkan loader is supplied by a working graphics driver. Avoid starting
	// the optional native backend when Windows would display a missing-DLL dialog.
	root := os.Getenv("SystemRoot")
	if filepath.IsAbs(root) && exists(filepath.Join(root, "System32", "vulkan-1.dll")) {
		return true
	}
	return exists(filepath.Join(filepath.Dir(e.path("llama-gpu", "llama-server")), "vulkan-1.dll"))
}
