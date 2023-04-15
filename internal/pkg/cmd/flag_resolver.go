package cmd

import (
	"fmt"
	"io"
	"os"

	"github.com/confluentinc/cli/internal/pkg/form"
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
	Prompt form.Prompt
	Out    io.Writer
}

// ValueFrom reads indirect flag values such as "-" for stdin pipe or "@file.txt" @ prefix
func (r *FlagResolverImpl) ValueFrom(source string, prompt string, secure bool) (string, error) {
	// Interactively prompt
	if source == "" {
		if prompt == "" {
			return "", ErrNoValueSpecified
		}

		if _, err := fmt.Fprint(r.Out, prompt); err != nil {
			return "", err
		}

		var value string
		var err error
		if secure {
			value, err = r.Prompt.ReadLineMasked()
		} else {
			value, err = r.Prompt.ReadLine()
		}
		if err != nil {
			return "", err
		}

		if _, err := fmt.Fprintf(r.Out, "\n"); err != nil {
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
		value, err := r.Prompt.ReadLine()
		if err != nil {
			return "", err
		}
		// To remove the final \n
		return value[0 : len(value)-1], nil
	}

	// Read from a file
	if source[0] == '@' {
		filePath := source[1:]
		b, err := os.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		return string(b), err
	}

	return source, nil
}
