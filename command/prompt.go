// go:generate mocker --prefix "" --out mock/prompt.go --pkg mock command Prompt

package command

import (
	"bufio"
	"fmt"
	"io"
	"os"

	"golang.org/x/crypto/ssh/terminal"
)

type Prompt interface {
	ReadString(delim byte) (string, error)
	ReadPassword(fd int) ([]byte, error)
	SetOutput(out io.Writer)
	Println(...interface{}) (n int, err error)
	Print(...interface{}) (n int, err error)
	Printf(format string, args ...interface{}) (n int, err error)
}

type TerminalPrompt struct {
	Stdin *bufio.Reader
	Out   io.Writer
}

func NewTerminalPrompt(reader io.Reader) Prompt {
	return &TerminalPrompt{Stdin: bufio.NewReader(reader), Out: os.Stdout}
}

func (p *TerminalPrompt) ReadString(delim byte) (string, error) {
	return p.Stdin.ReadString(delim)
}

func (p *TerminalPrompt) ReadPassword(fd int) ([]byte, error) {
	return terminal.ReadPassword(fd)
}

func (p *TerminalPrompt) SetOutput(out io.Writer) {
	p.Out = out
}

func (p *TerminalPrompt) Println(args ...interface{}) (n int, err error) {
	return fmt.Fprintln(p.Out, args...)
}

func (p *TerminalPrompt) Print(args ...interface{}) (n int, err error) {
	return fmt.Fprint(p.Out, args...)
}

func (p *TerminalPrompt) Printf(format string, args ...interface{}) (n int, err error) {
	return fmt.Fprintf(p.Out, format, args...)
}
