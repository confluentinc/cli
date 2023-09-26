package utils

import (
	"strings"
	"unicode"

	"github.com/spf13/pflag"
)

const (
	boolFlagType  = "bool"
	countFlagType = "count"
)

// Is flag that accepts argument
func IsFlagWithArg(flag *pflag.Flag) bool {
	if flag == nil {
		return false
	}
	flagType := flag.Value.Type()
	return flagType != boolFlagType && flagType != countFlagType
}

// count flag, e.g. verbose, may have its shorthand name repeated e.g. -vvvv
// The function returns true if the arg string is the shorthand of the count flag
func IsShorthandCountFlag(flag *pflag.Flag, arg string) bool {
	if flag.Value.Type() != countFlagType {
		return false
	}
	return arg == "-"+strings.Repeat(flag.Shorthand, len(arg)-1)
}

func IsFlagArg(arg string) bool {
	if len(arg) <= 1 {
		return false
	}
	if strings.HasPrefix(arg, "-") && unicode.IsLetter(rune(arg[1])) {
		return true
	}
	if len(arg) > 2 && strings.HasPrefix(arg, "--") && unicode.IsLetter(rune(arg[2])) {
		return true
	}
	return false
}
