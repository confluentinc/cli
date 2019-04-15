package mock

import (
	"bufio"
	"fmt"
)

// Prompt is a mock implementation of Prompt.
type Prompt struct {
	Strings       []string
	Passwords     []string
	StringIndex   int
	PasswordIndex int
	In            bufio.Reader
}

// ReadString returns the next string from Strings.
func (mock *Prompt) ReadString(delim byte) (string, error) {
	if len(mock.Strings) < mock.StringIndex {
		return "", fmt.Errorf("not enough mock strings")
	}
	mock.StringIndex++
	return mock.Strings[mock.StringIndex-1], nil
}

// ReadPassword returns the next password from Passwords
func (mock *Prompt) ReadPassword(fd int) ([]byte, error) {
	if len(mock.Passwords) < mock.PasswordIndex {
		return nil, fmt.Errorf("not enough mock strings")
	}
	mock.PasswordIndex++
	return []byte(mock.Passwords[mock.PasswordIndex-1]), nil
}
