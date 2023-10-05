package output

import (
	"fmt"
	"io"
	"os"
	"regexp"

	"github.com/charmbracelet/lipgloss"

	"github.com/confluentinc/cli/v3/pkg/resource"
)

var (
	codeSnippetRegexp = regexp.MustCompile("`[^`]+`")
	linkRegexp        = regexp.MustCompile(`https?://(www\.)?[a-z\-]+\.[/a-z\-]+`)
	resourceRegexp    = regexp.MustCompile(`"[^"]+"`)
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
		s = colorLinks(s)
		s = colorResources(s)
	}
	_, _ = fmt.Fprint(w, s)
}

func colorCodeSnippets(s string) string {
	return codeSnippetRegexp.ReplaceAllStringFunc(s, func(s string) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.Color("203")).Background(lipgloss.AdaptiveColor{Light: "254", Dark: "236"})
		return style.Render(s[1 : len(s)-1])
	})
}

func colorLinks(s string) string {
	return linkRegexp.ReplaceAllStringFunc(s, func(s string) string {
		style := lipgloss.NewStyle().Foreground(lipgloss.AdaptiveColor{Light: "36", Dark: "30"}).Underline(true)
		return style.Render(s)
	})
}

func colorResources(s string) string {
	return resourceRegexp.ReplaceAllStringFunc(s, func(s string) string {
		r := s[1 : len(s)-1]
		if resource.LookupType(r) == resource.Unknown {
			return s
		}

		style := lipgloss.NewStyle().Foreground(lipgloss.Color("#C69669"))
		return style.Render(r)
	})
}
