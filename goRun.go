package gorun

import (
	"io"
	"os/exec"
	"sync"
)

type GoRunConfig struct {
	ExecProgramPath string // eg: "server/main.exe"
	RunArguments    func() []string
	ExitChan        chan bool
	Logger          io.Writer
	KillAllOnStop   bool // If true, kills all instances of the executable when stopping
}

type GoRun struct {
	*GoRunConfig
	Cmd        *exec.Cmd
	isRunning  bool
	hasWaited  bool         // Track if Wait() has been called
	mutex      sync.RWMutex // Protect concurrent access to running state
	safeBuffer *SafeBuffer  // Thread-safe buffer for Logger
}

func New(c *GoRunConfig) *GoRun {
	var buffer *SafeBuffer
	if c.Logger != nil {
		// Create SafeBuffer that forwards to the original logger
		buffer = NewSafeBufferWithForward(c.Logger)
	} else {
		buffer = NewSafeBuffer()
	}

	return &GoRun{
		GoRunConfig: c,
		Cmd:         &exec.Cmd{},
		isRunning:   false,
		hasWaited:   false,
		mutex:       sync.RWMutex{},
		safeBuffer:  buffer,
	}
}

// GetOutput returns the captured output in a thread-safe manner
func (h *GoRun) GetOutput() string {
	return h.safeBuffer.String()
}
