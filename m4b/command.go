package m4b

import (
	"bytes"
	"os/exec"
)

type Cmd interface {
	Run(stdout *bytes.Buffer, stderr *bytes.Buffer) error
}

type Command interface {
	Create(name string, args ...string) Cmd
}

// ExecCommand is a wrapper for executing external commands
type ExecCommand struct{}

func (c *ExecCommand) Create(name string, args ...string) Cmd {
	return &ExecCmd{cmd: *exec.Command(name, args...)}
}

type ExecCmd struct {
	cmd exec.Cmd
}

func (c *ExecCmd) Run(stdout *bytes.Buffer, stderr *bytes.Buffer) error {
	c.cmd.Stderr = stderr
	c.cmd.Stdout = stdout
	return c.cmd.Run()
}
