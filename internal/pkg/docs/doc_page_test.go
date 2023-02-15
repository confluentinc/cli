package docs

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/examples"
)

func TestPrintDocPage(t *testing.T) {
	root1 := &cobra.Command{
		Use:   "a",
		Short: "Short description.",
	}

	root2 := &cobra.Command{
		Use:   "a",
		Short: "Short description.",
	}

	command1 := &cobra.Command{
		Use:   "b",
		Short: "Description of `b` 1.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `b` 1.",
				Code: "a b --flag 1",
			},
		),
		Run: doNothingFunc,
	}
	command1.Flags().String("flag", "", "Description of flag 1.")

	command2 := &cobra.Command{
		Use:   "b",
		Short: "Description of `b` 2.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `b` 2.",
				Code: "a b --flag 2",
			},
		),
		Run: doNothingFunc,
	}
	command2.Flags().String("flag", "", "Description of flag 2.")

	root1.AddCommand(command1)
	root2.AddCommand(command2)

	tabs := []Tab{
		{Name: "Tab 1", Command: command1},
		{Name: "Tab 2", Command: command2},
	}

	expected := []string{
		"..",
		"   WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.",
		"",
		".. _a_b:",
		"",
		"a b",
		"---",
		"",
		"Description",
		"~~~~~~~~~~~",
		"",
		".. tabs::",
		"",
		"   .. group-tab:: Tab 1",
		"   ",
		"      Description of ``b`` 1.",
		"      ",
		"      ::",
		"      ",
		"        a b [flags]",
		"      ",
		"   .. group-tab:: Tab 2",
		"   ",
		"      Description of ``b`` 2.",
		"      ",
		"      ::",
		"      ",
		"        a b [flags]",
		"      ",
		"Flags",
		"~~~~~",
		"",
		".. tabs::",
		"",
		"   .. group-tab:: Tab 1",
		"   ",
		"      ::",
		"      ",
		"            --flag string   Description of flag 1.",
		"        -h, --help          help for b",
		"      ",
		"   .. group-tab:: Tab 2",
		"   ",
		"      ::",
		"      ",
		"            --flag string   Description of flag 2.",
		"        -h, --help          help for b",
		"      ",
		"Examples",
		"~~~~~~~~",
		"",
		".. tabs::",
		"",
		"   .. group-tab:: Tab 1",
		"   ",
		"      Example of ``b`` 1.",
		"      ",
		"      ::",
		"      ",
		"        a b --flag 1",
		"      ",
		"   .. group-tab:: Tab 2",
		"   ",
		"      Example of ``b`` 2.",
		"      ",
		"      ::",
		"      ",
		"        a b --flag 2",
		"      ",
		"See Also",
		"~~~~~~~~",
		"",
		"* :ref:`a-ref` - Short description.",
		"",
	}

	require.Equal(t, expected, printDocPage(tabs, 1))
}

func TestPrintWarnings_NoWarnings(t *testing.T) {
	cmd := &cobra.Command{}
	require.Empty(t, printWarnings(cmd, 0))
}

func TestPrintWarnings_ConfluentLocal(t *testing.T) {
	confluent := &cobra.Command{Use: "confluent"}
	local := &cobra.Command{Use: "local"}
	confluent.AddCommand(local)

	expected := []string{
		".. include:: ../includes/cli.rst",
		"  :end-before: cli_limitations_end",
		"  :start-after: cli_limitations_start",
		"",
	}

	require.Equal(t, expected, printWarnings(local, 1))
}

func TestPrintSphinxBlock_NoArgs(t *testing.T) {
	expected := []string{
		".. key:: val",
		"",
	}

	require.Equal(t, expected, printSphinxBlock("key", "val", nil))
}

func TestPrintSphinxBlock_Args(t *testing.T) {
	args := map[string]string{
		"key-a": "val-a",
		"key-b": "val-b",
	}

	expected := []string{
		".. key:: val",
		"  :key-a: val-a",
		"  :key-b: val-b",
		"",
	}

	require.Equal(t, expected, printSphinxBlock("key", "val", args))
}

func TestPrintUsageAndDescription(t *testing.T) {
	cmd := &cobra.Command{
		Use:  "command",
		Long: "Description of `command`.",
	}

	expected := []string{
		"Description of ``command``.",
		"",
		"::",
		"",
		"  command [flags]",
		"",
	}

	rows, ok := printDescriptionAndUsage(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintNotes_NoNotes(t *testing.T) {
	cmd := &cobra.Command{}
	require.Empty(t, printNotes(cmd, 0))
}

func TestPrintNotes_ConfluentLocal(t *testing.T) {
	confluent := &cobra.Command{Use: "confluent"}
	local := &cobra.Command{Use: "local"}
	confluent.AddCommand(local)

	expected := []string{
		".. include:: ../includes/path-set-cli.rst",
		"",
	}

	require.Equal(t, expected, printNotes(local, 1))
}

func TestPrintNotes_ConfluentIAMRoleBindingCreate(t *testing.T) {
	confluent := &cobra.Command{Use: "confluent"}
	iam := &cobra.Command{Use: "iam"}
	rbac := &cobra.Command{Use: "rbac"}
	roleBinding := &cobra.Command{Use: "role-binding"}
	create := &cobra.Command{Use: "create"}

	confluent.AddCommand(iam)
	iam.AddCommand(rbac)
	rbac.AddCommand(roleBinding)
	roleBinding.AddCommand(create)

	expected := []string{
		".. note:: If you need to troubleshoot when setting up role bindings, it may be helpful to view audit logs on the fly to identify authorization events for specific principals, resources, or operations. For details, refer to :platform:`Viewing audit logs on the fly|security/audit-logs/audit-logs-properties-config.html#view-audit-logs-on-the-fly`.",
		"",
	}

	require.Equal(t, expected, printNotes(create, 3))
}

func TestPrintNotes_ConfluentSecret(t *testing.T) {
	confluent := &cobra.Command{Use: "confluent"}
	secret := &cobra.Command{Use: "secret"}
	confluent.AddCommand(secret)

	expected := []string{
		".. tip:: For examples, see :platform:`Secrets Usage Examples|security/secrets.html#secrets-examples`.",
		"",
	}

	require.Equal(t, expected, printNotes(secret, 1))
}

func TestPrintFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("flag", "", "Flag description.")

	expected := []string{
		"::",
		"",
		"      --flag string   Flag description.\n",
	}

	section, ok := printFlags(cmd)
	require.True(t, ok)
	require.Equal(t, expected, section)
}

func TestPrintFlags_RequiredFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("flag", "", "Flag description.")
	require.NoError(t, cmd.MarkFlagRequired("flag"))

	expected := []string{
		"::",
		"",
		"      --flag string   REQUIRED: Flag description.\n",
	}

	section, ok := printFlags(cmd)
	require.True(t, ok)
	require.Equal(t, expected, section)
}

func TestPrintGlobalFlags(t *testing.T) {
	a := &cobra.Command{}
	a.PersistentFlags().String("global-flag", "", "Global flag description.")
	b := &cobra.Command{}

	a.AddCommand(b)

	expected := []string{
		"::",
		"",
		"      --global-flag string   Global flag description.\n",
	}

	rows, ok := printGlobalFlags(b)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintFlagSet_NoFlags(t *testing.T) {
	section, ok := printFlagSet(new(pflag.FlagSet))
	require.False(t, ok)
	require.Equal(t, []string{"No flags.", ""}, section)
}

func TestPrintFlagSet(t *testing.T) {
	flags := new(pflag.FlagSet)
	flags.String("flag", "", "Flag description.")

	expected := []string{
		"::",
		"",
		"      --flag string   Flag description.\n",
	}

	rows, ok := printFlagSet(flags)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintExamplesSection_NoExamples(t *testing.T) {
	cmd := &cobra.Command{}

	rows, ok := printExamples(cmd)
	require.False(t, ok)
	require.Equal(t, []string{"No examples.", ""}, rows)
}

func TestPrintExamplesSection_TextOnly(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Text: "Text-only example."},
		),
	}

	expected := []string{
		"Text-only example.",
		"",
	}

	rows, ok := printExamples(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintExamplesSection_CodeOnly(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Code: "command subcommand"},
		),
	}

	expected := []string{
		"::",
		"",
		"  command subcommand",
		"",
	}

	rows, ok := printExamples(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintExamplesSection_DoubleCodeBlock(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Code: "command subcommand"},
			examples.Example{Code: "command subcommand"},
		),
	}

	expected := []string{
		"::",
		"",
		"  command subcommand",
		"  command subcommand",
		"",
	}

	rows, ok := printExamples(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintExamplesSection_FullExample(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `command subcommand`.",
				Code: "command subcommand",
			},
		),
	}

	expected := []string{
		"Example of ``command subcommand``.",
		"",
		"::",
		"",
		"  command subcommand",
		"",
	}

	rows, ok := printExamples(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintExamplesSection_TwoExamples(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `command subcommand`.",
				Code: "command subcommand --flag 1",
			},
			examples.Example{
				Text: "Another example of `command subcommand`.",
				Code: "command subcommand --flag 2",
			},
		),
	}

	expected := []string{
		"Example of ``command subcommand``.",
		"",
		"::",
		"",
		"  command subcommand --flag 1",
		"",
		"Another example of ``command subcommand``.",
		"",
		"::",
		"",
		"  command subcommand --flag 2",
		"",
	}

	rows, ok := printExamples(cmd)
	require.True(t, ok)
	require.Equal(t, expected, rows)
}

func TestPrintSeeAlso(t *testing.T) {
	a := &cobra.Command{Use: "a", Short: "Description of A."}
	b := &cobra.Command{Use: "b"}

	a.AddCommand(b)

	expected := []string{
		"* :ref:`a-ref` - Description of A.",
		"",
	}

	require.Equal(t, expected, printSeeAlso(b))
}
