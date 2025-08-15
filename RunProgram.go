package gorun

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

func (h *GoRun) RunProgram() error {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	// Always stop any previous running program first
	// Use cleanup if KillAllOnStop is enabled
	if h.KillAllOnStop {
		if err := h.stopProgramAndCleanupUnsafe(true); err != nil {
			fmt.Fprintf(h.safeBuffer, "Warning: Error stopping previous programs: %v\n", err)
		}
	} else {
		if err := h.stopProgramUnsafe(); err != nil {
			fmt.Fprintf(h.safeBuffer, "Warning: Error stopping previous program: %v\n", err)
		}
	}

	runArgs := []string{}

	if h.RunArguments != nil {
		runArgs = h.RunArguments()
	}

	h.Cmd = exec.Command(h.ExecProgramPath, runArgs...)
	h.hasWaited = false // Reset wait flag for new process

	stderr, err := h.Cmd.StderrPipe()
	if err != nil {
		return err
	}

	stdout, err := h.Cmd.StdoutPipe()
	if err != nil {
		return err
	}

	err = h.Cmd.Start()
	if err != nil {
		return err
	}

	h.isRunning = true

	var once sync.Once
	done := make(chan struct{})

	// Create local references for goroutines to avoid race conditions
	currentCmd := h.Cmd

	go io.Copy(h.safeBuffer, stderr)
	go io.Copy(h.safeBuffer, stdout)

	go func() {
		select {
		case <-h.ExitChan:
			// h.Print("Received exit signal, stopping application...")
			h.StopProgram()
			once.Do(func() { close(done) })
		case <-done:
			// finish goroutine
		}
	}()

	go func() {
		err := currentCmd.Wait()
		h.mutex.Lock()
		h.isRunning = false
		h.hasWaited = true // Mark that Wait() has been called
		h.mutex.Unlock()

		if err != nil {
			fmt.Fprintf(h.safeBuffer, "App: %v closed with error: %v\n", h.ExecProgramPath, err)
		} else {
			fmt.Fprintf(h.safeBuffer, "App: %v closed successfully\n", h.ExecProgramPath)
		}
		once.Do(func() { close(done) })
	}()

	return nil
}
