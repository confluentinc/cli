package errors

import (
	"fmt"
)

func New(msg string) error {
	return fmt.Errorf(msg)
}

func Wrap(err error, msg string) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", msg, err)
}

func Wrapf(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return fmt.Errorf("%s: %w", fmt.Sprintf(format, args...), err)
}

func Errorf(format string, args ...any) error {
	return fmt.Errorf(format, args...)
}

func CustomMultierrorList(errors []error) string {
	if len(errors) == 1 {
		return errors[0].Error()
	}

	errString := fmt.Sprintf("%d errors occurred:", len(errors))
	for _, err := range errors {
		errString = fmt.Sprintf("%s\n\t* %v", errString, err)
	}
	return errString
}
