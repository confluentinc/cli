package errors

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

func CheckNoOpUpdate(flags *pflag.FlagSet, flagsToCheck ...string) error {
	var unsetFlags []string
	for _, flag := range flagsToCheck {
		if !flags.Changed(flag) {
			unsetFlags = append(unsetFlags, fmt.Sprintf("`--%s`", flag))
		}
	}
	if len(unsetFlags) == len(flagsToCheck) {
		return errors.New(fmt.Sprintf("one of the following flags must be set: %s", strings.Join(unsetFlags, ", ")))
	}
	return nil
}
