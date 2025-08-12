package gorun

import (
	"io"
	"os/exec"
)

type GoRunConfig struct {
	ExecProgramPath string // eg: "server/main.exe"
	RunArguments    func() []string
	ExitChan        chan bool
	Writer          io.Writer
}

type GoRun struct {
	*GoRunConfig
	Cmd       *exec.Cmd
	IsRunning bool
}

func New(c *GoRunConfig) *GoRun {
	return &GoRun{
		GoRunConfig: c,
		Cmd:         &exec.Cmd{},
		IsRunning:   false,
	}
}
