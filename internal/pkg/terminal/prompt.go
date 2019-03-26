package terminal

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

// Prompt represents input and output to a terminal
type Prompt interface {
	ReadString(delim byte) (string, error)
	ReadPassword(fd int) ([]byte, error)
	GetOutput() io.Writer
	SetOutput(out io.Writer)
	Println(...interface{}) (n int, err error)
	Print(...interface{}) (n int, err error)
	Printf(format string, args ...interface{}) (n int, err error)
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
func (p *RealPrompt) ReadPassword(fd int) ([]byte, error) {
	return terminal.ReadPassword(fd)
}

// GetOutput returns the writer to which the Print* methods will write.
func (p *RealPrompt) GetOutput() io.Writer {
	return p.Out
}

// SetOutput updates the writer to which the Print* methods will write.
func (p *RealPrompt) SetOutput(out io.Writer) {
	p.Out = out
}

// Println formats using the default formats for its operands and writes to Out.
// Spaces are always added between operands and a newline is appended.
// It returns the number of bytes written and any write error encountered.
func (p *RealPrompt) Println(args ...interface{}) (n int, err error) {
	return fmt.Fprintln(p.Out, args...)
}

// Print formats using the default formats for its operands and writes to Out.
// Spaces are added between operands when neither is a string.
// It returns the number of bytes written and any write error encountered.
func (p *RealPrompt) Print(args ...interface{}) (n int, err error) {
	return fmt.Fprint(p.Out, args...)
}

// Printf formats according to a format specifier and writes to Out.
// It returns the number of bytes written and any write error encountered.
func (p *RealPrompt) Printf(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(p.Out, format, args...)
}
