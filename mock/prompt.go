package mock

import (
	"bufio"
	"fmt"
	"io"
)

// Prompt is a mock implementation of Prompt.
type Prompt struct {
	Strings       []string
	Passwords     []string
	StringIndex   int
	PasswordIndex int
	In            bufio.Reader
	Out           io.Writer
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

// GetOutput gets the output writer
func (mock *Prompt) GetOutput() io.Writer {
	return mock.Out
}

// SetOutput sets the output writer
func (mock *Prompt) SetOutput(out io.Writer) {
	mock.Out = out
}

// Println calls fmt.Println
func (mock *Prompt) Println(args ...interface{}) (int, error) {
	return fmt.Fprintln(mock.Out, args...)
}

// Print calls fmt.Print
func (mock *Prompt) Print(args ...interface{}) (int, error) {
	return fmt.Fprint(mock.Out, args...)
}

// Printf calls fmt.Printf
func (mock *Prompt) Printf(format string, args ...interface{}) (int, error) {
	return fmt.Fprintf(mock.Out, format, args...)
}
