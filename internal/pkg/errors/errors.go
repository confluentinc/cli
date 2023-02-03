package errors

import (
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
