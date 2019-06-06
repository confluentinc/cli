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
	IsPipe() (bool, error)
}

// RealPrompt is the standard prompt implementation
type RealPrompt struct {
	In    *bufio.Reader
	Out   io.Writer
	Stdin *os.File
}

// NewPrompt returns a new RealPrompt instance which reads from reader and writes to Stdout.
func NewPrompt(stdin *os.File) *RealPrompt {
	return &RealPrompt{In: bufio.NewReader(stdin), Out: os.Stdout, Stdin: stdin}
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
func (p *RealPrompt) ReadString(delim byte) (string, error) {
	return p.In.ReadString(delim)
}

// ReadPassword reads a line of input from a terminal without local echo.
func (p *RealPrompt) ReadPassword() ([]byte, error) {
	return gopass.GetPasswd()
}

// ReadPassword reads a line of input from a terminal without local echo.
func (p *RealPrompt) IsPipe() (bool, error) {
	fi, err := p.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (fi.Mode() & os.ModeCharDevice) == 0, nil
}
