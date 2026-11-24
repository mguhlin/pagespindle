//go:build windows

package engine

import (
	"os/exec"
	"syscall"
)

// Bundled helpers are single processes. No shell or helper-spawn API is exposed.
func configureProcess(c *exec.Cmd) {
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: 0x08000000}
}
func killProcess(c *exec.Cmd) error {
	if c.Process == nil {
		return nil
	}
	return c.Process.Kill()
}
