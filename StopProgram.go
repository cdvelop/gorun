package gorun

import (
	"fmt"
	"os"
	"runtime"
	"syscall"
	"time"
)

func (h *GoRun) StopProgram() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if h.KillAllOnStop {
		return h.stopProgramAndCleanupUnsafe(true)
	}
	return h.stopProgramUnsafe()
}

// stopProgramUnsafe stops the program without acquiring the mutex
// Should only be called when mutex is already held
func (h *GoRun) stopProgramUnsafe() error {
	if !h.isRunning || h.Cmd.Process == nil {
		h.isRunning = false
		return nil
	}

	// Check if process has already exited
	if h.Cmd.ProcessState != nil && h.Cmd.ProcessState.Exited() {
		h.isRunning = false
		return nil
	}

	process := h.Cmd.Process
	h.isRunning = false

	// Cross-platform graceful shutdown approach
	if runtime.GOOS == "windows" {
		// On Windows, we don't have SIGTERM, so we use Kill directly
		return process.Kill()
	}

	// On Unix-like systems (Linux, macOS), try graceful shutdown first
	if err := process.Signal(syscall.SIGTERM); err != nil {
		// If SIGTERM fails, force kill
		return process.Kill()
	}

	// Wait a bit for graceful shutdown
	done := make(chan error, 1)
	go func() {
		_, err := process.Wait()
		done <- err
	}()

	select {
	case <-time.After(3 * time.Second):
		// Timeout reached, force kill
		fmt.Fprintf(os.Stderr, "Process did not terminate gracefully, forcing kill\n")
		return process.Kill()
	case err := <-done:
		// Process terminated gracefully
		return err
	}
}
