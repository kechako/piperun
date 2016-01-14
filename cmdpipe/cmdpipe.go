package cmdpipe

import (
	"errors"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type CmdPipe struct {
	Name       string
	Args       []string
	Cmd        *exec.Cmd
	PipeReader *io.PipeReader
	PipeWriter *io.PipeWriter
	ExitStatus int
	Error      error
}

func NewCmdPipe() *CmdPipe {
	return &CmdPipe{
		Args: make([]string, 0, 10),
	}
}

func (c *CmdPipe) CreateCmd() {
	c.Cmd = exec.Command(c.Name, c.Args...)
}

func (c *CmdPipe) Start() error {
	if c.Cmd == nil {
		c.CreateCmd()
	}

	if c.PipeReader != nil {
		c.Cmd.Stdin = c.PipeReader
	} else {
		c.Cmd.Stdin = os.Stdin
	}

	if c.PipeWriter != nil {
		c.Cmd.Stdout = c.PipeWriter
	} else {
		c.Cmd.Stdout = os.Stdout
	}

	c.Cmd.Stderr = os.Stderr

	return c.Cmd.Start()
}

func (c *CmdPipe) Wait() {
	if c.PipeWriter != nil {
		defer func() {
			err := c.PipeWriter.Close()
			if err != nil && c.Error == nil {
				c.ExitStatus = 4
				c.Error = err
			}
		}()
	}
	err := c.Cmd.Wait()
	if err != nil {
		c.Error = err
		if exitErr, ok := err.(*exec.ExitError); ok {
			if status, ok := exitErr.Sys().(syscall.WaitStatus); ok {
				c.ExitStatus = status.ExitStatus()
			} else {
				panic(errors.New("Unimplemented for system where exec.ExitError.Sys() is not syscall.WaitStatus."))
			}
		} else {
			c.ExitStatus = 3
		}
	}
}
