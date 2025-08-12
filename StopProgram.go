package gorun

func (h *GoRun) StopProgram() error {
	if !h.IsRunning || h.Cmd.Process == nil {
		h.IsRunning = false
		return nil
	}

	// First try to find if process exists
	if h.Cmd.ProcessState != nil && h.Cmd.ProcessState.Exited() {
		h.IsRunning = false
		return nil
	}

	h.IsRunning = false
	return h.Cmd.Process.Kill()
}
