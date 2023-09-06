package errors

import (
	"fmt"

	"github.com/pkg/errors"
)

func New(msg string) error {
	return errors.New(msg)
}

func Wrap(err error, msg string) error {
	return errors.Wrap(err, msg)
}

func Wrapf(err error, fmt string, args ...any) error {
	return errors.Wrapf(err, fmt, args...)
}

func Errorf(fmt string, args ...any) error {
	return errors.Errorf(fmt, args...)
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
