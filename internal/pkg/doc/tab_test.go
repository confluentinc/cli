package doc

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrintTabbedPages(t *testing.T) {
	tabs := []Tab{
		{Name: "Title 1"},
		{Name: "Title 2"},
	}

	pages := []string{
		"Content 1\n",
		"Content 2\n",
	}

	expected := strings.Join([]string{
		".. tabs::",
		"",
		"   .. tab:: Title 1",
		"",
		"      Content 1",
		"      ",
		"   .. tab:: Title 2",
		"",
		"      Content 2",
		"      ",
	}, "\n")

	require.Equal(t, expected, printTabbedPages(tabs, pages))
}

func TestPrintTabbedPage(t *testing.T) {
	expected := strings.Join([]string{
		"   .. tab:: Title",
		"",
		"      Content",
		"      ",
	}, "\n")

	require.Equal(t, expected, printTabbedPage("Title", "Content\n"))
}

func TestIndent(t *testing.T) {
	require.Equal(t, " a\n b", indent(" ", "a\nb"))
}
