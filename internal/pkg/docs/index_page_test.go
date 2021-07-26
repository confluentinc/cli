package docs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var (
	printHeaderFunc = func(_ *cobra.Command) []string { return []string{"HEADER", ""} }
	doNothingFunc   = func(_ *cobra.Command, _ []string) {}
)

func TestPrintIndexPage(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}

	a1 := &cobra.Command{Use: "a", Short: "Description 1."}
	a2 := &cobra.Command{Use: "a", Short: "Description 2."}

	cmd.AddCommand(a1)
	cmd.AddCommand(a2)

	a1.AddCommand(&cobra.Command{Use: "b1", Short: "Description 1.", Run: doNothingFunc})
	a2.AddCommand(&cobra.Command{Use: "b2", Short: "Description 2.", Run: doNothingFunc})

	tabs := []Tab{
		{Name: "Tab 1", Command: a1},
		{Name: "Tab 2", Command: a2},
	}

	expected := []string{
		".. _command_a:",
		"",
		"command a",
		"=========",
		"",
		".. toctree::",
		"   :hidden:",
		"",
		"   command_a_b1",
		"   command_a_b2",
		"",
		"Description",
		"~~~~~~~~~~~",
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
		"Subcommands",
		"~~~~~~~~~~~",
		"",
		".. tabs::",
		"",
		"   .. group-tab:: Tab 1",
		"   ",
		"      ===================== ================",
		"             Command          Description   ",
		"      ===================== ================",
		"       :ref:`command_a_b1`   Description 1. ",
		"      ===================== ================",
		"      ",
		"   .. group-tab:: Tab 2",
		"   ",
		"      ===================== ================",
		"             Command          Description   ",
		"      ===================== ================",
		"       :ref:`command_a_b2`   Description 2. ",
		"      ===================== ================",
		"      ",
	}

	require.Equal(t, expected, printIndexPage(tabs))
}

func TestFlatten(t *testing.T) {
	arrs := [][]string{
		{"a", "b"},
		{"c", "d"},
	}

	require.Equal(t, []string{"a", "b", "c", "d"}, flatten(arrs))
}

func TestPrintHeader(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		".. _command-ref:",
		"",
	}

	require.Equal(t, expected, printHeader(cmd))
}

func TestPrintTitle_Root(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		"|confluent| CLI Command Reference",
		"=================================",
		"",
	}

	require.Equal(t, expected, printTitle(cmd, "="))
}

func TestPrintTitle_NonRoot(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	expected := []string{
		"a b",
		"---",
		"",
	}

	require.Equal(t, expected, printTitle(b, "-"))
}

func TestPrintTableOfContents(t *testing.T) {
	a1 := &cobra.Command{Use: "a"}
	a2 := &cobra.Command{Use: "a"}

	b1 := &cobra.Command{Use: "b1", Run: doNothingFunc}
	b2 := &cobra.Command{Use: "b2", Run: doNothingFunc}

	a1.AddCommand(b1)
	a2.AddCommand(b2)

	tabs := []Tab{
		{Name: "Tab 1", Command: a1},
		{Name: "Tab 2", Command: a2},
	}

	expected := []string{
		".. toctree::",
		"   :hidden:",
		"",
		"   a_b1",
		"   a_b2",
		"",
	}

	require.Equal(t, expected, printTableOfContents(tabs))
}

func TestPrintLink(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	require.Equal(t, "a/index", printLink(a))
	require.Equal(t, "a_b", printLink(b))
}

func TestPrintLongestDescription_Short(t *testing.T) {
	cmd := &cobra.Command{Short: "Description."}
	require.Equal(t, "Description.", printLongestDescription(cmd))
}

func TestPrintLongestDescription_Long(t *testing.T) {
	cmd := &cobra.Command{
		Short: "Description.",
		Long:  "Long description.",
	}
	require.Equal(t, "Long description.", printLongestDescription(cmd))
}

func TestFormatReST(t *testing.T) {
	require.Equal(t, "Description of ``command``.", formatReST("Description of `command`."))
}

func TestPrintSubcommands(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{
		Use:   "b",
		Short: "Short description.",
		Run:   doNothingFunc,
	}

	a.AddCommand(b)

	expected := []string{
		"============ ====================",
		"  Command        Description     ",
		"============ ====================",
		" :ref:`a_b`   Short description. ",
		"============ ====================",
		"",
	}

	rows, ok := printSubcommands(a)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintSphinxRef(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}
	require.Equal(t, ":ref:`command-ref`", printSphinxRef(cmd))
}

func TestPrintRef_Root(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}
	require.Equal(t, "command-ref", printRef(cmd))
}

func TestPrintRef(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	require.Equal(t, "a_b", printRef(b))
}

func TestDedent(t *testing.T) {
	arr := []string{
		" a ",
		" b ",
	}

	require.Equal(t, []string{"a", "b"}, dedent(arr))
}
