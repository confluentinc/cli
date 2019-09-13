package cmd

import (
	"fmt"
	"io"
	"io/ioutil"
)

var (
	ErrUnexpectedStdinPipe = fmt.Errorf("unexpected stdin pipe")
	ErrNoValueSpecified    = fmt.Errorf("no value specified")
	ErrNoPipe              = fmt.Errorf("no pipe")
)

// FlagResolver reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
type FlagResolver interface {
	ValueFrom(source string, prompt string, secure bool) (string, error)
}

type FlagResolverImpl struct {
	Prompt Prompt
	Out    io.Writer
}

// ValueFrom reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
func (r *FlagResolverImpl) ValueFrom(source string, prompt string, secure bool) (value string, err error) {
	// Interactively prompt
	if source == "" {
		if prompt == "" {
			return "", ErrNoValueSpecified
		}
		if yes, err := r.Prompt.IsPipe(); err != nil {
			return "", err
		} else if yes {
			return "", ErrUnexpectedStdinPipe
		}

		_, err = fmt.Fprintf(r.Out, prompt)
		if err != nil {
			return "", err
		}
		if secure {
			valueByte, err := r.Prompt.ReadPassword()
			if err != nil {
				return "", err
			}
			value = string(valueByte)
		} else {
			value, err = r.Prompt.ReadString('\n')
		}
		_, err = fmt.Fprintf(r.Out, "\n")
		if err != nil {
			return "", err
		}
		return value, err
	}

	// Read from stdin pipe
	if source == "-" {
		if yes, err := r.Prompt.IsPipe(); err != nil {
			return "", err
		} else if !yes {
			return "", ErrNoPipe
		}
		value, err = r.Prompt.ReadString('\n')
		if err != nil {
			return "", err
		}
		// To remove the final \n
		return value[0 : len(value)-1], nil
	}

	// Read from a file
	if source[0] == '@' {
		filePath := source[1:]
		b, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(b), err
	}

	return source, nil
}
