// go:generate mocker --prefix "" --out mock/prompt.go --pkg mock command Prompt

package command

import (
	"bufio"
	"io"

	"golang.org/x/crypto/ssh/terminal"
)

type Prompt interface {
	ReadString(delim byte) (string, error)
	ReadPassword(fd int) ([]byte, error)
}

type TerminalPrompt struct {
	Stdin *bufio.Reader
}

func NewTerminalPrompt(reader io.Reader) Prompt {
	return &TerminalPrompt{Stdin: bufio.NewReader(reader)}
}

func (p TerminalPrompt) ReadString(delim byte) (string, error) {
	return p.Stdin.ReadString(delim)
}

func (p TerminalPrompt) ReadPassword(fd int) ([]byte, error) {
	return terminal.ReadPassword(fd)
}
