package doc

import (
	"strings"

	"github.com/spf13/cobra"
)

// Tab represents a tab on a documentation page, for showing commands of the same name with differing details.
type Tab struct {
	Name    string
	Command *cobra.Command
}

func printTabbedPages(tabs []Tab, pages []string) string {
	isUnified := true
	page := pages[0]

	for i := 1; i < len(pages); i++ {
		if page != pages[i] {
			isUnified = false
		}
	}

	if !isUnified {
		rows := []string{
			".. tabs::",
			"",
		}

		for i, page := range pages {
			rows = append(rows, printTabbedPage(tabs[i].Name, page))
		}

		page = strings.Join(rows, "\n")
	}

	return page
}

func printTabbedPage(name, page string) string {
	tab := "   "
	rows := []string{
		tab + ".. tab:: " + name,
		"",
		indent(tab+tab, page),
	}
	return strings.Join(rows, "\n")
}

func indent(tab, s string) string {
	rows := strings.Split(s, "\n")
	for i := range rows {
		rows[i] = tab + rows[i]
	}
	return strings.Join(rows, "\n")
}
