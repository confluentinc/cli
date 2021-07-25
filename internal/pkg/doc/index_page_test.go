package doc

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var (
	printHeaderFunc = func(_ *cobra.Command) []string { return []string{"HEADER", ""} }
	doNothingFunc   = func(_ *cobra.Command, _ []string) {}
)

func TestPrintTabbedIndexPage(t *testing.T) {
	a := &cobra.Command{Use: "command"}
	b := &cobra.Command{Use: "command"}

	a.AddCommand(&cobra.Command{Use: "subcommand", Short: "Description A", Run: doNothingFunc})
	b.AddCommand(&cobra.Command{Use: "subcommand", Short: "Description B", Run: doNothingFunc})

	tabs := []Tab{
		{Name: "Tab A", Command: a},
		{Name: "Tab B", Command: b},
	}

	expected := strings.Join([]string{
		".. tabs::",
		"",
		"   .. tab:: Tab A",
		"",
		"      HEADER",
		"      ",
		"      .. toctree::",
		"         :hidden:",
		"      ",
		"         command_subcommand",
		"      ",
		"      =========================== ===============",
		"                Command             Description  ",
		"      =========================== ===============",
		"       :ref:`command_subcommand`   Description A ",
		"      =========================== ===============",
		"      ",
		"   .. tab:: Tab B",
		"",
		"      HEADER",
		"      ",
		"      .. toctree::",
		"         :hidden:",
		"      ",
		"         command_subcommand",
		"      ",
		"      =========================== ===============",
		"                Command             Description  ",
		"      =========================== ===============",
		"       :ref:`command_subcommand`   Description B ",
		"      =========================== ===============",
		"      ",
	}, "\n")

	require.Equal(t, expected, printTabbedIndexPage(tabs, printHeaderFunc))
}

func TestPrintIndexPage(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{
		Use:   "b",
		Short: "A subcommand with subcommands.",
	}
	c := &cobra.Command{
		Use:   "c",
		Short: "A subcommand with no subcommands.",
		Run:   doNothingFunc,
	}

	a.AddCommand(b)
	a.AddCommand(c)
	b.AddCommand(&cobra.Command{Run: doNothingFunc})

	expected := strings.Join([]string{
		"HEADER",
		"",
		".. toctree::",
		"   :hidden:",
		"",
		"   b/index",
		"   a_c",
		"",
		"============ ===================================",
		"  Command                Description            ",
		"============ ===================================",
		" :ref:`a_b`   A subcommand with subcommands.    ",
		" :ref:`a_c`   A subcommand with no subcommands. ",
		"============ ===================================",
		"",
	}, "\n")

	require.Equal(t, expected, printIndexPage(a, printHeaderFunc))
}

func TestPrintIndexHeader_ShortDescription(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Short description.",
	}

	expected := []string{
		".. _command-ref:",
		"",
		"command",
		"=======",
		"",
		"Description",
		"~~~~~~~~~~~",
		"",
		"Short description.",
		"",
	}

	require.Equal(t, expected, printIndexHeader(cmd))
}

func TestPrintIndexHeader_LongDescription(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Short description.",
		Long:  "Long description.",
	}

	expected := []string{
		".. _command-ref:",
		"",
		"command",
		"=======",
		"",
		"Description",
		"~~~~~~~~~~~",
		"",
		"Long description.",
		"",
	}

	require.Equal(t, expected, printIndexHeader(cmd))
}

func TestPrintTableOfContents(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}
	c := &cobra.Command{Use: "c", Run: doNothingFunc}

	a.AddCommand(b)
	a.AddCommand(c)
	b.AddCommand(&cobra.Command{Run: doNothingFunc})

	expected := []string{
		".. toctree::",
		"   :hidden:",
		"",
		"   b/index",
		"   a_c",
		"",
	}

	require.Equal(t, expected, printTableOfContents(a))
}

func TestPrintSubcommands(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{
		Use:   "b",
		Short: "Description of b.",
		Run:   doNothingFunc,
	}

	a.AddCommand(b)

	expected := []string{
		"============ ===================",
		"  Command        Description    ",
		"============ ===================",
		" :ref:`a_b`   Description of b. ",
		"============ ===================",
		"",
	}

	require.Equal(t, expected, printSubcommands(a))
}

func TestPrintLink(t *testing.T) {
	a := &cobra.Command{Use: "a"}
	b := &cobra.Command{Use: "b"}
	c := &cobra.Command{Use: "c"}

	a.AddCommand(b)
	b.AddCommand(c)

	require.Equal(t, "a/index", printLink(a))
	require.Equal(t, "b/index", printLink(b))
	require.Equal(t, "a_b_c", printLink(c))
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
