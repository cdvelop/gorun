package gorun

import (
	"fmt"
	"io"
	"os/exec"
	"sync"
)

func (h *GoRun) RunProgram() error {

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
	h.IsRunning = true

	var once sync.Once
	done := make(chan struct{})

	go io.Copy(h.Writer, stderr)
	go io.Copy(h.Writer, stdout)

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
		if err != nil {
			fmt.Fprintf(h.Writer, "App: %v closed with error: %v\n", h.ExecProgramPath, err)
		} else {
			fmt.Fprintf(h.Writer, "App: %v closed successfully\n", h.ExecProgramPath)
		}
		once.Do(func() { close(done) })
	}()

	return nil
}
