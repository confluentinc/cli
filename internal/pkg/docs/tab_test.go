package docs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestPrintTabbedSection_Hidden(t *testing.T) {
	t.Parallel()

	printSection := func(*cobra.Command) ([]string, bool) { return []string{}, false }
	tabs := make([]Tab, 1)

	rows := printTabbedSection("", printSection, tabs)
	require.Empty(t, rows)
}

func TestPrintTabbedSection_Unified(t *testing.T) {
	t.Parallel()

	printSection := func(*cobra.Command) ([]string, bool) { return []string{"Content"}, true }
	tabs := make([]Tab, 2)

	expected := []string{
		"Title",
		"~~~~~",
		"",
		"Content",
	}

	require.Equal(t, expected, printTabbedSection("Title", printSection, tabs))
}

func TestPrintTabbedSection_Tabbed(t *testing.T) {
	t.Parallel()

	printSection := func(cmd *cobra.Command) ([]string, bool) { return []string{cmd.Short, ""}, true }
	tabs := []Tab{
		{
			Name:    "Tab 1",
			Command: &cobra.Command{Short: "Description 1."},
		},
		{
			Name:    "Tab 2",
			Command: &cobra.Command{Short: "Description 2."},
		},
	}

	expected := []string{
		"Title",
		"~~~~~",
		"",
		".. tabs::",
		"",
		"   .. group-tab:: Tab 1",
		"   ",
		"      Description 1.",
		"      ",
		"   .. group-tab:: Tab 2",
		"   ",
		"      Description 2.",
		"      ",
	}

	require.Equal(t, expected, printTabbedSection("Title", printSection, tabs))
}

func TestPrintSection(t *testing.T) {
	t.Parallel()

	expected := []string{
		"Title",
		"~~~~~",
		"",
		"Line 1",
		"Line 2",
	}

	require.Equal(t, expected, printSection("Title", []string{"Line 1", "Line 2"}))
}

func TestAreEqual_DifferentLen(t *testing.T) {
	t.Parallel()

	require.False(t, areEqual([]string{}, []string{""}))
}

func TestAreEqual_DifferentElements(t *testing.T) {
	t.Parallel()

	require.False(t, areEqual([]string{"a"}, []string{"b"}))
}

func TestAreEqual_True(t *testing.T) {
	t.Parallel()

	require.True(t, areEqual([]string{"a"}, []string{"a"}))
}

func TestIndent(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{" a", " b"}, indent(" ", []string{"a", "b"}))
}

func TestIndent_WithNewlines(t *testing.T) {
	t.Parallel()

	require.Equal(t, []string{" a", " b", " c"}, indent(" ", []string{"a", "b\nc"}))
}
