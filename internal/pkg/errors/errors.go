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

func Wrapf(err error, fmt string, args ...interface{}) error {
	return errors.Wrapf(err, fmt, args...)
}

func Errorf(fmt string, args ...interface{}) error {
	return errors.Errorf(fmt, args...)
}

func Cause(err error) error {
	return errors.Cause(err)
}
