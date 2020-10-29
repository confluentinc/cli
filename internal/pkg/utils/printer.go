package utils

import (
	"fmt"

	"github.com/spf13/cobra"
)

//
// These printers are needed because we want to write to stdout, not stderr
// as (cobra.Command).Print* does by default. So we also add ErrPrint* too.
//

// Println formats using the default formats for its operands and writes to stdout.
// Spaces are always added between operands and a newline is appended.
func Println(cmd *cobra.Command, args ...interface{}) {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), args...)
}

// Print formats using the default formats for its operands and writes to stdout.
// Spaces are added between operands when neither is a string.
func Print(cmd *cobra.Command, args ...interface{}) {
	_, _ = fmt.Fprint(cmd.OutOrStdout(), args...)
}

// Printf formats according to a format specifier and writes to stdout.
func Printf(cmd *cobra.Command, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(cmd.OutOrStdout(), format, args...)
}

// ErrPrintln formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands and a newline is appended.
func ErrPrintln(cmd *cobra.Command, args ...interface{}) {
	_, _ = fmt.Fprintln(cmd.OutOrStderr(), args...)
}

// ErrPrint formats using the default formats for its operands and writes to stderr.
// Spaces are added between operands when neither is a string.
func ErrPrint(cmd *cobra.Command, args ...interface{}) {
	_, _ = fmt.Fprint(cmd.OutOrStderr(), args...)
}

// ErrPrintf formats according to a format specifier and writes to stderr.
func ErrPrintf(cmd *cobra.Command, format string, args ...interface{}) {
	_, _ = fmt.Fprintf(cmd.OutOrStderr(), format, args...)
}
