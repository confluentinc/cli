package errors

import (
	"errors"
	"fmt"
)

func New(msg string) error {
	return errors.New(msg)
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
