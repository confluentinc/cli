package exec

import (
	"os/exec"
)

type Command interface {
	Output() ([]byte, error)
}

type RealCommand struct {
	Cmd *exec.Cmd
}

func NewCommand(name string, args ...string) *RealCommand {
	return &RealCommand{Cmd: exec.Command(name, args...)}
}

func (c *RealCommand) Output() ([]byte, error) {
	return c.Cmd.Output()
}
