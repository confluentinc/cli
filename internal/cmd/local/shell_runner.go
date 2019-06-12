package local

import (
	"io"

	"github.com/DABH/go-basher"
)

// ShellRunner interface used to run shell scripts
type ShellRunner interface {
	Init(stdout io.Writer, stderr io.Writer)
	Export(name string, value string)
	Source(filepath string, loader func(string) ([]byte, error)) error
	Run(command string, args []string) (int, error)
}

// BashShellRunner is an implementation of ShellRunner using go-basher
type BashShellRunner struct {
	BasherContext *basher.Context
}

// Init initializes a runner's output streams
func (runner *BashShellRunner) Init(stdout io.Writer, stdin io.Writer) {
	runner.BasherContext.Stdout = stdout
	runner.BasherContext.Stderr = stdin
}

// Export a key=value pair for the shell being used
func (runner *BashShellRunner) Export(name string, value string) {
	runner.BasherContext.Export(name, value)
}

// Source source a script for the shell
func (runner *BashShellRunner) Source(filepath string, loader func(string) ([]byte, error)) error {
	return runner.BasherContext.Source(filepath, loader)
}

// Run a command within the shell
func (runner *BashShellRunner) Run(command string, args []string) (int, error) {
	return runner.BasherContext.Run(command, args)
}
