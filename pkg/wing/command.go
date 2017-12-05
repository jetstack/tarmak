// Copyright Jetstack Ltd. See LICENSE for details.
package wing

import (
	"io"
	"os"
	"os/exec"
)

// command interface to make it more testable
type Command interface {
	Start() error
	StderrPipe() (io.ReadCloser, error)
	StdinPipe() (io.WriteCloser, error)
	StdoutPipe() (io.ReadCloser, error)
	Wait() error
	Process() *os.Process
}

// execCommand wraps a real command
var _ Command = &execCommand{}

type execCommand struct {
	*exec.Cmd
}

func (e *execCommand) Process() *os.Process {
	return e.Cmd.Process
}
