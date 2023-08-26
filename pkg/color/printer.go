package color

import (
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

func Print(color bool, s string) {
	printTo(os.Stdout, color, s)
}

func Println(color bool, s string) {
	printTo(os.Stdout, color, s+"\n")
}

func Printf(color bool, s string, args ...any) {
	printTo(os.Stdout, color, fmt.Sprintf(s, args...))
}

func ErrPrint(color bool, s string) {
	printTo(os.Stderr, color, s)
}

func ErrPrintln(color bool, s string) {
	printTo(os.Stderr, color, s+"\n")
}

func ErrPrintf(color bool, s string, args ...any) {
	printTo(os.Stderr, color, fmt.Sprintf(s, args...))
}

func printTo(w io.Writer, color bool, s string) {
	if color {
		s = colorCodeSnippets(s)
		s = colorErrors(s)
		s = colorLinks(s)
	}
	_, _ = fmt.Fprint(w, s)
}

func colorCodeSnippets(s string) string {
	re := regexp.MustCompile("`[^`]+`")
	return re.ReplaceAllStringFunc(s, func(s string) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"})
		return style.Render(s[1 : len(s)-1])
	})
}

func colorErrors(s string) string {
	if strings.HasPrefix(s, "Error: ") {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("203"))
		return style.Render(s)
	}
	return s
}

func colorLinks(s string) string {
	re := regexp.MustCompile(`https?://(www\.)?[a-z\-]+\.[/a-z\-]+`)
	return re.ReplaceAllStringFunc(s, func(s string) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "36", Dark: "30"}).Underline(true)
		return style.Render(s)
	})
}
