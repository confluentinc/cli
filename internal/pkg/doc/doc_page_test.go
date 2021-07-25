package doc

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/internal/pkg/examples"
)

func TestPrintDocPage(t *testing.T) {
	cmd := &cobra.Command{
		Use:   "command",
		Short: "Command description.",
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `command`.",
				Code: "command",
			},
		),
	}
	cmd.Flags().String("flag", "", "Description of flag.")
	subcommand := &cobra.Command{
		Use:   "subcommand",
		Short: "Description of subcommand.",
		Run:   doNothingFunc,
	}

	cmd.AddCommand(subcommand)

	expected := strings.Join([]string{
		".. _command-ref:",
		"",
		"command",
		"-------",
		"",
		"Description",
		"~~~~~~~~~~~",
		"",
		"Command description.",
		"",
		"::",
		"",
		"  command [flags]",
		"",
		"Flags",
		"~~~~~",
		"",
		"::",
		"",
		"      --flag string   Description of flag.",
		"",
		"Examples",
		"~~~~~~~~",
		"",
		"Example of ``command``.",
		"",
		"::",
		"",
		"  command",
		"",
		"See Also",
		"~~~~~~~~",
		"",
		"* :ref:`command_subcommand` - Description of subcommand.",
		"",
	}, "\n")

	require.Equal(t, expected, printDocPage(cmd, 0))
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
		"command",
		"-------",
		"",
	}

	require.Equal(t, expected, printHeader(cmd, "-"))
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

func TestPrintDescription(t *testing.T) {
	cmd := &cobra.Command{Short: "Command description."}

	expected := []string{
		"Description",
		"~~~~~~~~~~~",
		"",
		"Command description.",
		"",
	}

	require.Equal(t, expected, printDescription(cmd))
}

func TestPrintUsage(t *testing.T) {
	cmd := &cobra.Command{Use: "command"}

	expected := []string{
		"::",
		"",
		"  command",
		"",
	}

	require.Equal(t, expected, printUsage(cmd))
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

func TestPrintNotes_ConfluentIAMRolebindingCreate(t *testing.T) {
	confluent := &cobra.Command{Use: "confluent"}
	iam := &cobra.Command{Use: "iam"}
	rolebinding := &cobra.Command{Use: "rolebinding"}
	create := &cobra.Command{Use: "create"}

	confluent.AddCommand(iam)
	iam.AddCommand(rolebinding)
	rolebinding.AddCommand(create)

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

func TestPrintFlags_RequiredFlag(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("flag", "", "Flag description.")
	require.NoError(t, cmd.MarkFlagRequired("flag"))

	expected := []string{
		"Flags",
		"~~~~~",
		"",
		"::",
		"",
		"      --flag string   REQUIRED: Flag description.\n",
	}

	require.Equal(t, expected, printFlags(cmd))
}

func TestPrintFlags(t *testing.T) {
	a := &cobra.Command{}
	a.PersistentFlags().String("global-flag", "", "Global flag description.")

	b := &cobra.Command{}
	b.Flags().String("local-flag", "", "Local flag description.")

	a.AddCommand(b)

	expected := []string{
		"Flags",
		"~~~~~",
		"",
		"::",
		"",
		"      --local-flag string   Local flag description.\n",
		"Global Flags",
		"~~~~~~~~~~~~",
		"",
		"::",
		"",
		"      --global-flag string   Global flag description.\n",
	}

	require.Equal(t, expected, printFlags(b))
}

func TestPrintFlagSet_NoFlags(t *testing.T) {
	require.Empty(t, printFlagSet("", &pflag.FlagSet{}))
}

func TestPrintFlagSet(t *testing.T) {
	flags := &pflag.FlagSet{}
	flags.String("flag", "", "Flag description.")

	expected := []string{
		"Flags",
		"~~~~~",
		"",
		"::",
		"",
		"      --flag string   Flag description.\n",
	}

	require.Equal(t, expected, printFlagSet("Flags", flags))
}

func TestPrintExamples_NoExamples(t *testing.T) {
	cmd := &cobra.Command{}
	require.Empty(t, printExamples(cmd))
}

func TestPrintExamples_TextOnly(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Text: "Text-only example."},
		),
	}

	expected := []string{
		"Examples",
		"~~~~~~~~",
		"",
		"Text-only example.",
		"",
	}

	require.Equal(t, expected, printExamples(cmd))
}

func TestPrintExamples_TextWithCodeSnippet(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Text: "Text with a `code snippet`."},
		),
	}

	expected := []string{
		"Examples",
		"~~~~~~~~",
		"",
		"Text with a ``code snippet``.",
		"",
	}

	require.Equal(t, expected, printExamples(cmd))
}

func TestPrintExamples_CodeOnly(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Code: "command subcommand"},
		),
	}

	expected := []string{
		"Examples",
		"~~~~~~~~",
		"",
		"::",
		"",
		"  command subcommand",
		"",
	}

	require.Equal(t, expected, printExamples(cmd))
}

func TestPrintExamples_DoubleCodeBlock(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{Code: "command subcommand"},
			examples.Example{Code: "command subcommand"},
		),
	}

	expected := []string{
		"Examples",
		"~~~~~~~~",
		"",
		"::",
		"",
		"  command subcommand",
		"  command subcommand",
		"",
	}

	require.Equal(t, expected, printExamples(cmd))
}

func TestPrintExamples_FullExample(t *testing.T) {
	cmd := &cobra.Command{
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Example of `command subcommand`.",
				Code: "command subcommand",
			},
		),
	}

	expected := []string{
		"Examples",
		"~~~~~~~~",
		"",
		"Example of ``command subcommand``.",
		"",
		"::",
		"",
		"  command subcommand",
		"",
	}

	require.Equal(t, expected, printExamples(cmd))
}

func TestPrintExamples_TwoExamples(t *testing.T) {
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
		"Examples",
		"~~~~~~~~",
		"",
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

	require.Equal(t, expected, printExamples(cmd))
}

func TestSeeAlso(t *testing.T) {
	a := &cobra.Command{Use: "a", Short: "Description of A."}
	b := &cobra.Command{Use: "b"}
	c := &cobra.Command{Use: "c", Short: "Description of C.", Run: doNothingFunc}

	a.AddCommand(b)
	b.AddCommand(c)

	expected := []string{
		"See Also",
		"~~~~~~~~",
		"",
		"* :ref:`a-ref` - Description of A.",
		"* :ref:`a_b_c` - Description of C.",
		"",
	}

	require.Equal(t, expected, printSeeAlso(b))
}
