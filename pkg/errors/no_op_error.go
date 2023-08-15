package errors

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
)

func CheckNoUpdate(flags *pflag.FlagSet, flagsToCheck ...string) error {
	var unsetFlags []string
	for _, flag := range flagsToCheck {
		if !flags.Changed(flag) {
			unsetFlags = append(unsetFlags, fmt.Sprintf("`--%s`", flag))
		}
	}
	if len(unsetFlags) == len(flagsToCheck) {
		return fmt.Errorf("at least one of the following flags must be set: %s", strings.Join(unsetFlags, ", "))
	}
	return nil
}
