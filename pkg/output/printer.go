package output

import (
	"fmt"
	"io"
	"os"
)

//
// These printers are needed because we want to write to stdout, not stderr
// as (cobra.Command).Print* does by default. So we also add ErrPrint* too.
//

// Print formats using the default formats for its operands and writes to stdout.
// Spaces are added between operands when neither is a string.
func Print(s string) {
	printTo(os.Stdout, s)
}

// Println formats using the default formats for its operands and writes to stdout.
// Spaces are always added between operands and a newline is appended.
func Println(s string) {
	printTo(os.Stdout, s+"\n")
}

// Printf formats according to a format specifier and writes to stdout.
func Printf(s string, args ...any) {
	printTo(os.Stdout, fmt.Sprintf(s, args...))
}

// ErrPrint formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands.
func ErrPrint(s string) {
	printTo(os.Stderr, s)
}

// ErrPrintln formats using the default formats for its operands and writes to stderr.
// Spaces are always added between operands and a newline is appended.
func ErrPrintln(s string) {
	printTo(os.Stderr, s+"\n")
}

// ErrPrintf formats according to a format specifier and writes to stderr.
func ErrPrintf(s string, args ...any) {
	printTo(os.Stderr, fmt.Sprintf(s, args...))
}

func printTo(w io.Writer, s string) {
	_, _ = fmt.Fprint(w, s)
}
