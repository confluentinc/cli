package errors

import (
	"errors"

	perrors "github.com/pkg/errors"
)

func New(msg string) error {
	return perrors.New(msg)
}

func Wrap(err error, msg string) error {
	return perrors.Wrap(err, msg)
}

func Wrapf(err error, fmt string, args ...any) error {
	return perrors.Wrapf(err, fmt, args...)
}

func Errorf(fmt string, args ...any) error {
	return perrors.Errorf(fmt, args...)
}

func Join(errs ...error) error {
	return errors.Join(errs...)
}
