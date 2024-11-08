package m4b

import (
	"bytes"
	"os/exec"
)

// Cmd represents an executable command that can be run with stdout/stderr capture
type Cmd interface {
	Run(stdout, stderr *bytes.Buffer) error
	RunI(stdin *bytes.Reader, stdout, stderr *bytes.Buffer) error
}

// Command is a factory interface for creating executable commands
type Command interface {
	Create(name string, args ...string) Cmd
}

// ExecCommand is a wrapper for executing external commands that implements the Command interface
type ExecCommand struct{}

// Create returns a new Cmd instance that will execute the specified command with given arguments
func (c *ExecCommand) Create(name string, args ...string) Cmd {
	return &ExecCmd{cmd: *exec.Command(name, args...)}
}

// ExecCmd is a wrapper for exec.Cmd for testability
type ExecCmd struct {
	cmd exec.Cmd
}

// Run executes the command and captures its output in the provided stdout and stderr buffers
func (c *ExecCmd) Run(stdout *bytes.Buffer, stderr *bytes.Buffer) error {
	c.cmd.Stderr = stderr
	c.cmd.Stdout = stdout
	return c.cmd.Run()
}

// RunI works like Run, except that it also takes a stdin
func (c *ExecCmd) RunI(stdin *bytes.Reader, stdout, stderr *bytes.Buffer) error {
	c.cmd.Stdin = stderr
	c.cmd.Stderr = stderr
	c.cmd.Stdout = stdout
	return c.cmd.Run()
}
