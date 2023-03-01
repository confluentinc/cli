package utils

import (
	"fmt"
	"os"
)

//
// These printers are needed because we want to write to stdout, not stderr
// as (cobra.Command).Print* does by default. So we also add ErrPrint* too.
//

// Print formats using the default formats for its operands and writes to stdout.
// Spaces are added between operands when neither is a string.
func Print(args ...any) {
	_, _ = fmt.Fprint(os.Stdout, args...)
}

// Println formats using the default formats for its operands and writes to stdout.
// Spaces are always added between operands and a newline is appended.
func Println(args ...any) {
	_, _ = fmt.Fprintln(os.Stdout, args...)
}

// Printf formats according to a format specifier and writes to stdout.
func Printf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// ErrPrint formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands.
func ErrPrint(args ...any) {
	_, _ = fmt.Fprint(os.Stderr, args...)
}

// ErrPrintln formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands and a newline is appended.
func ErrPrintln(args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, args...)
}

// ErrPrintf formats according to a format specifier and writes to stderr.
func ErrPrintf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
