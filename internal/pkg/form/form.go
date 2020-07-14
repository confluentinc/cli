package form

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"syscall"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

/*
A multi-question form. Examples:

* Signup
First Name: Brian
Last Name: Strauch
Email: bstrauch@confluent.io

* Login
Username: user
Password: ****

* Confirmation
Submit? (y/n): y

* Defaults
Save File As: (file.txt) other.txt
*/

type Form struct {
	Fields    []Field
	Responses map[string]interface{}
}

type Field struct {
	ID           string
	Prompt       string
	DefaultValue interface{}
	IsYesOrNo    bool
	IsHidden     bool
}

func New(fields ...Field) *Form {
	return &Form{
		Fields:    fields,
		Responses: make(map[string]interface{}),
	}
}

func (f *Form) Prompt(in *bufio.Reader, out *bufio.Writer) error {
	for _, field := range f.Fields {
		if err := show(out, field, f.Responses[field.ID]); err != nil {
			return err
		}

		val, err := read(in, field)
		if err != nil {
			return err
		}

		res, err := save(field, val)
		if err != nil {
			return err
		}
		f.Responses[field.ID] = res
	}

	return nil
}

func show(out *bufio.Writer, field Field, savedValue interface{}) error {
	line := field.Prompt

	if field.IsYesOrNo {
		line += " (y/n)"
	}
	line += ": "

	if savedValue != nil {
		line += fmt.Sprintf("(%v) ", savedValue)
	} else if field.DefaultValue != nil {
		line += fmt.Sprintf("(%v) ", field.DefaultValue)
	}

	if _, err := out.WriteString(line); err != nil {
		return err
	}
	return out.Flush()
}

func read(in *bufio.Reader, field Field) (string, error) {
	if field.IsHidden {
		if err := setTTY("-echo"); err != nil {
			return "", err
		}
	}

	val, err := in.ReadString('\n')
	val = strings.TrimSuffix(val, "\n")
	if err == io.EOF {
		err = nil
	}

	if field.IsHidden {
		if err := setTTY("echo"); err != nil {
			return "", err
		}
	}

	return val, err
}

func setTTY(opt string) error {
	attrs := syscall.ProcAttr{
		Files: []uintptr{os.Stdin.Fd(), os.Stdout.Fd(), os.Stderr.Fd()},
	}

	pid, err := syscall.ForkExec("/bin/stty", []string{"stty", opt}, &attrs)
	if err != nil {
		return err
	}

	_, err = syscall.Wait4(pid, new(syscall.WaitStatus), 0, nil)
	return err
}

func save(field Field, val string) (interface{}, error) {
	if field.IsYesOrNo {
		switch strings.ToUpper(val) {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}
		return false, fmt.Errorf(errors.InvalidChoiceMsg, val)
	}

	if val == "" {
		return field.DefaultValue, nil
	}

	return val, nil
}
