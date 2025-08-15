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
	Cmd       *exec.Cmd
	isRunning bool
	mutex     sync.RWMutex // Protect concurrent access to running state
}

func New(c *GoRunConfig) *GoRun {
	return &GoRun{
		GoRunConfig: c,
		Cmd:         &exec.Cmd{},
		isRunning:   false,
		mutex:       sync.RWMutex{},
	}
}
