//go:build !windows

package engine

func (e *Engine) gpuRuntimeAvailable() bool { return true }
