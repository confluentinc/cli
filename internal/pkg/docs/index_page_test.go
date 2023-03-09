package docs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var doNothingFunc = func(_ *cobra.Command, _ []string) {}

func TestPrintIndexPage(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}

	a1 := &cobra.Command{Use: "a", Short: "Description 1.", Aliases: []string{"alias"}}
	a2 := &cobra.Command{Use: "a", Short: "Description 2.", Aliases: []string{"alias"}}

	cmd.AddCommand(a1)
	cmd.AddCommand(a2)

	a1.AddCommand(&cobra.Command{Use: "b1", Short: "Description 1.", Run: doNothingFunc})
	a2.AddCommand(&cobra.Command{Use: "b2", Short: "Description 2.", Run: doNothingFunc})

	tabs := []Tab{
		{Name: "Tab 1", Command: a1},
		{Name: "Tab 2", Command: a2},
	}

	expected := []string{
		"..",
		"   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.",
		"",
		".. _command_a:",
		"",
		"command a",
		"=========",
		"",
		"Aliases",
		"~~~~~~~",
		"",
		"::",
		"",
		"  a, alias",
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
		".. toctree::",
		"   :hidden:",
		"",
		"   command_a_b1",
		"   command_a_b2",
		"",
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

	require.Equal(t, expected, printIndexPage(tabs, false))
}

func TestPrintRootIndexPage(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}

	a1 := &cobra.Command{Use: "a"}
	a2 := &cobra.Command{Use: "a"}

	cmd.AddCommand(a1)
	cmd.AddCommand(a2)

	a1.AddCommand(&cobra.Command{Use: "b1", Run: doNothingFunc})
	a2.AddCommand(&cobra.Command{Use: "b2", Run: doNothingFunc})

	tabs := []Tab{
		{Name: "Tab 1", Command: a1},
		{Name: "Tab 2", Command: a2},
	}

	expected := []string{
		"..",
		"   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.",
		"",
		".. _command_a:",
		"",
		"command a",
		"=========",
		"",
		".. raw:: html",
		"",
		`   <script type="text/javascript">`,
		"      window.location = 'overview.html';",
		"   </script>",
		"",
		".. toctree::",
		"   :hidden:",
		"",
		"   command_a_b1",
		"   command_a_b2",
		"",
	}

	require.Equal(t, expected, printRootIndexPage(tabs))
}

func TestFlatten(t *testing.T) {
	t.Parallel()

	arrs := [][]string{
		{"a", "b"},
		{"c", "d"},
	}

	require.Equal(t, []string{"a", "b", "c", "d"}, flatten(arrs))
}

func TestPrintComments(t *testing.T) {
	t.Parallel()

	expected := []string{
		"..",
		"   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.",
		"",
	}

	require.Equal(t, expected, printComments())
}

func TestPrintHeader(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		".. _command-ref:",
		"",
	}

	require.Equal(t, expected, printHeader(cmd, false))
}

func TestPrintTitle_Root(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		"|confluent| CLI Command Reference",
		"=================================",
		"",
	}

	require.Equal(t, expected, printTitle(cmd, "="))
}

func TestPrintTitle_NonRoot(t *testing.T) {
	t.Parallel()

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

func TestPrintInlineScript(t *testing.T) {
	t.Parallel()

	expected := []string{
		".. raw:: html",
		"",
		`   <script type="text/javascript">`,
		"      window.location = 'overview.html';",
		"   </script>",
		"",
	}

	require.Equal(t, expected, printInlineScript())
}

func TestPrintTableOfContents(t *testing.T) {
	t.Parallel()

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
		"   :maxdepth: 1",
		"   :hidden:",
		"",
		"   Overview <overview>",
		"   a_b1",
		"   a_b2",
		"",
	}

	require.Equal(t, expected, printTableOfContents(tabs))
}

func TestPrintLink(t *testing.T) {
	t.Parallel()

	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	require.Equal(t, "a/index", printLink(a))
	require.Equal(t, "a_b", printLink(b))
}

func TestPrintAliases_Empty(t *testing.T) {
	t.Parallel()

	cmd := new(cobra.Command)
	require.Empty(t, printAliases(cmd))
}

func TestPrintAliases(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Use:     "long-command",
		Aliases: []string{"lc"},
	}

	expected := []string{
		"::",
		"",
		"  long-command, lc",
		"",
	}

	require.Equal(t, expected, printAliases(cmd))
}

func TestPrintDescription_Root(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		"The available |confluent| CLI commands are documented here.",
		"",
	}

	actual, ok := printDescription(cmd)
	require.True(t, ok)
	require.Equal(t, expected, actual)
}

func TestPrintDescription(t *testing.T) {
	t.Parallel()

	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b", Short: "Description."}

	a.AddCommand(b)

	expected := []string{
		"Description.",
		"",
	}

	actual, ok := printDescription(b)
	require.True(t, ok)
	require.Equal(t, expected, actual)
}

func TestPrintLongestDescription_Short(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Short: "Description."}
	require.Equal(t, "Description.", printLongestDescription(cmd))
}

func TestPrintLongestDescription_Long(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{
		Short: "Description.",
		Long:  "Long description.",
	}
	require.Equal(t, "Long description.", printLongestDescription(cmd))
}

func TestFormatReST_CodeSnippet(t *testing.T) {
	t.Parallel()

	require.Equal(t, "Description of ``command``.", formatReST("Description of `command`."))
}

func TestFormatReST_Target(t *testing.T) {
	t.Parallel()

	require.Equal(t, `"target\_" "target\_"`, formatReST(`"target_" "target_"`))
}

func TestPrintSubcommands(t *testing.T) {
	t.Parallel()

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
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}
	require.Equal(t, ":ref:`command-ref`", printSphinxRef(cmd))
}

func TestPrintRef_Root(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}
	require.Equal(t, "command-ref", printRef(cmd, false))
}

func TestPrintRef_Overview(t *testing.T) {
	t.Parallel()

	cmd := &cobra.Command{Use: "command"}
	require.Equal(t, "command-ref-index", printRef(cmd, true))
}

func TestPrintRef(t *testing.T) {
	t.Parallel()

	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	require.Equal(t, "a_b", printRef(b, false))
}

func TestDedent(t *testing.T) {
	t.Parallel()

	arr := []string{
		" a ",
		" b ",
	}

	require.Equal(t, []string{"a", "b"}, dedent(arr))
}
