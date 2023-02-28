package utils

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

//
// These printers are needed because we want to write to stdout, not stderr
// as (cobra.Command).Print* does by default. So we also add ErrPrint* too.
//

// Println formats using the default formats for its operands and writes to stdout.
// Spaces are always added between operands and a newline is appended.
func Println(cmd *cobra.Command, args ...any) {
	_, _ = fmt.Fprintln(os.Stdout, args...)
}

// Print formats using the default formats for its operands and writes to stdout.
// Spaces are added between operands when neither is a string.
func Print(cmd *cobra.Command, args ...any) {
	_, _ = fmt.Fprint(os.Stdout, args...)
}

// Printf formats according to a format specifier and writes to stdout.
func Printf(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stdout, format, args...)
}

// ErrPrint formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands.
func ErrPrint(cmd *cobra.Command, args ...any) {
	_, _ = fmt.Fprint(os.Stderr, args...)
}

// ErrPrintln formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands and a newline is appended.
func ErrPrintln(cmd *cobra.Command, args ...any) {
	_, _ = fmt.Fprintln(os.Stderr, args...)
}

// ErrPrintf formats according to a format specifier and writes to stderr.
func ErrPrintf(cmd *cobra.Command, format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, format, args...)
}
