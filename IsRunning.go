package gorun

// IsRunning returns whether the program is currently running
func (h *GoRun) IsRunning() bool {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return h.isRunning
}

// GetPID returns the process ID if the program is running, otherwise returns 0
func (h *GoRun) GetPID() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	if h.isRunning && h.Cmd.Process != nil {
		return h.Cmd.Process.Pid
	}
	return 0
}
