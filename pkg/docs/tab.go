package docs

import (
	"strings"

	"github.com/spf13/cobra"
)

// Tab represents a tab on a documentation page, for showing commands of the same name with differing details.
type Tab struct {
	Name    string
	Command *cobra.Command
}

func printTabbedSection(title string, printSectionFunc func(*cobra.Command) ([]string, bool), tabs []Tab) []string {
	sections := make([][]string, len(tabs))
	isHidden := true

	for i, tab := range tabs {
		section, ok := printSectionFunc(tab.Command)
		sections[i] = section
		if ok {
			isHidden = false
		}
	}

	if isHidden {
		return []string{}
	}

	isUnified := true
	for i := 1; i < len(sections); i++ {
		if !areEqual(sections[0], sections[i]) {
			isUnified = false
		}
	}

	rows := sections[0]

	if !isUnified {
		rows = []string{
			".. tabs::",
			"",
		}

		for i, tab := range tabs {
			section := []string{
				".. group-tab:: " + tab.Name,
				"",
			}
			section = append(section, indent("   ", sections[i])...)

			rows = append(rows, indent("   ", section)...)
		}
	}

	return printSection(title, rows)
}

func printSection(title string, section []string) []string {
	if len(section) == 0 {
		return []string{}
	}

	head := []string{
		title,
		strings.Repeat("~", len(title)),
		"",
	}

	return append(head, section...)
}

func areEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func indent(tab string, rows []string) []string {
	var indented []string
	for _, row := range rows {
		for _, line := range strings.Split(row, "\n") {
			indented = append(indented, tab+line)
		}
	}
	return indented
}
