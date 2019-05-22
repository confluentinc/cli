package cmd

import (
	"bufio"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// Prompt represents input and output to a terminal
type Prompt interface {
	ReadString(delim byte) (string, error)
	ReadPassword() (string, error)
}

// RealPrompt is the standard prompt implementation
type RealPrompt struct {
	Stdin   *bufio.Reader
	Out     io.Writer
	StdinFD int
}

// NewPrompt returns a new RealPrompt instance which reads from reader and writes to Stdout.
func NewPrompt(stdin *os.File) *RealPrompt {
	return &RealPrompt{Stdin: bufio.NewReader(stdin), Out: os.Stdout, StdinFD: int(stdin.Fd())}
}

// ReadString reads until the first occurrence of delim in the input,
// returning a string containing the data up to and including the delimiter.
func (p *RealPrompt) ReadString(delim byte) (string, error) {
	return p.Stdin.ReadString(delim)
}

// ReadPassword reads a line of input from a terminal without local echo.
func (p *RealPrompt) ReadPassword() (string, error) {
	pass, err := terminal.ReadPassword(p.StdinFD)
	if err != nil {
		return "", err
	}
	return string(pass), nil
}
