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
			fmt.Fprintf(h.Logger, "Warning: Error stopping previous programs: %v\n", err)
		}
	} else {
		if err := h.stopProgramUnsafe(); err != nil {
			fmt.Fprintf(h.Logger, "Warning: Error stopping previous program: %v\n", err)
		}
	}

	runArgs := []string{}

	if h.RunArguments != nil {
		runArgs = h.RunArguments()
	}

	h.Cmd = exec.Command(h.ExecProgramPath, runArgs...)

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

	go io.Copy(h.Logger, stderr)
	go io.Copy(h.Logger, stdout)

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
		err := h.Cmd.Wait()
		h.mutex.Lock()
		h.isRunning = false
		h.mutex.Unlock()

		if err != nil {
			fmt.Fprintf(h.Logger, "App: %v closed with error: %v\n", h.ExecProgramPath, err)
		} else {
			fmt.Fprintf(h.Logger, "App: %v closed successfully\n", h.ExecProgramPath)
		}
		once.Do(func() { close(done) })
	}()

	return nil
}
