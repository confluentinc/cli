//go:generate go run github.com/travisjeffery/mocker/cmd/mocker --prefix "" --dst ../mock/prompt.go --pkg mock --selfpkg github.com/confluentinc/cli/v4 prompt.go Prompt

package form

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/havoc-io/gopass"
)

// Prompt represents input and output to a terminal
type Prompt interface {
	ReadLine() (string, error)
	ReadLineMasked() (string, error)
	IsPipe() (bool, error)
}

// RealPrompt is the standard prompt implementation
type RealPrompt struct {
	In    *bufio.Reader
	Out   io.Writer
	Stdin *os.File
}

// NewPrompt returns a new RealPrompt instance which reads from reader and writes to Stdout.
func NewPrompt() *RealPrompt {
	return &RealPrompt{In: bufio.NewReader(os.Stdin), Out: os.Stdout, Stdin: os.Stdin}
}

// ReadLine reads a line of input, without the newline.
func (p *RealPrompt) ReadLine() (string, error) {
	str, err := p.In.ReadString('\n')
	return strings.TrimRight(str, "\r\n"), err
}

// ReadLineMasked reads a line of input from a terminal without local echo.
func (p *RealPrompt) ReadLineMasked() (string, error) {
	isPipe, err := p.IsPipe()
	if err != nil {
		return "", err
	}
	if isPipe {
		return p.ReadLine()
	}

	pwd, err := gopass.GetPasswdMasked()
	if err != nil && err.Error() == "interrupted" {
		_, _ = fmt.Fprintln(p.Out, "^C")
		os.Exit(0)
	}
	return string(pwd), err
}

func (p *RealPrompt) IsPipe() (bool, error) {
	fi, err := p.Stdin.Stat()
	if err != nil {
		return false, err
	}
	return (fi.Mode() & os.ModeCharDevice) == 0, nil
}
