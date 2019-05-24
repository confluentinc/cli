package mock

import (
	"bufio"
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

// Prompt is a mock implementation of Prompt.
type Prompt struct {
	Strings       []string
	Passwords     []string
	StringIndex   int
	PasswordIndex int
	Pipe          bool
	In            bufio.Reader
}

var _ pcmd.Prompt = (*Prompt)(nil)

// ReadString returns the next string from Strings.
func (mock *Prompt) ReadString(delim byte) (string, error) {
	if len(mock.Strings) < mock.StringIndex {
		return "", fmt.Errorf("not enough mock strings")
	}
	mock.StringIndex++
	return mock.Strings[mock.StringIndex-1], nil
}

// ReadPassword returns the next password from Passwords
func (mock *Prompt) ReadPassword() (string, error) {
	if len(mock.Passwords) < mock.PasswordIndex {
		return "", fmt.Errorf("not enough mock strings")
	}
	mock.PasswordIndex++
	return mock.Passwords[mock.PasswordIndex-1], nil
}

func (mock *Prompt) IsPipe() (bool, error) {
	return mock.Pipe, nil
}
