package cmd

import (
	"bufio"
	"io"
	"os"

	"github.com/havoc-io/gopass"
)

// Prompt represents input and output to a terminal
type Prompt interface {
	ReadString(delim byte) (string, error)
	ReadPassword() ([]byte, error)
}

// RealPrompt is the standard prompt implementation
type RealPrompt struct {
	Stdin *bufio.Reader
	Out   io.Writer
}

// NewPrompt returns a new RealPrompt instance which reads from reader and writes to Stdout.
func NewPrompt(reader io.Reader) *RealPrompt {
	return &RealPrompt{Stdin: bufio.NewReader(reader), Out: os.Stdout}
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
func (p *RealPrompt) ReadString(delim byte) (string, error) {
	return p.Stdin.ReadString(delim)
}

// ReadPassword reads a line of input from a terminal without local echo.
func (p *RealPrompt) ReadPassword() ([]byte, error) {
	return gopass.GetPasswd()
}
