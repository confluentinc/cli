package docs

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/types"
)

// generateDocPage creates a file which contains the command description, usage, flags, examples, and more.
// If there are multiple versions of a single command, tabs are used within the file.
func generateDocPage(tabs []Tab, dir string, depth int) error {
	name := strings.ReplaceAll(tabs[0].Command.CommandPath(), " ", "_")
	path := filepath.Join(dir, name+".rst")
	rows := printDocPage(tabs, depth)

	return writeFile(path, strings.Join(rows, "\n"))
}

func printDocPage(tabs []Tab, depth int) []string {
	cmd := tabs[0].Command

	return flatten([][]string{
		printComments(),
		printHeader(cmd, false),
		printTitle(cmd, "-"),
		printWarnings(cmd, depth),
		printTabbedSection("Description", printDescriptionAndUsage, tabs),
		printNotes(cmd, depth),
		printTabbedSection("Flags", printFlags, tabs),
		printTabbedSection("Global Flags", printGlobalFlags, tabs),
		printTabbedSection("Examples", printExamples, tabs),
		printSection("See Also", printSeeAlso(cmd)),
	})
}

func printWarnings(cmd *cobra.Command, depth int) []string {
	var rows []string

	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		include := strings.Repeat("../", depth) + "includes/cli.rst"
		args := map[string]string{
			"start-after": "cli_limitations_start",
			"end-before":  "cli_limitations_end",
		}
		rows = append(rows, printSphinxBlock("include", include, args)...)
	}

	return rows
}

func printSphinxBlock(key, val string, args map[string]string) []string {
	rows := []string{
		fmt.Sprintf(".. %s:: %s", key, val),
	}

	keys := types.GetSortedKeys(args)

	for _, key := range keys {
		rows = append(rows, fmt.Sprintf("  :%s: %s", key, args[key]))
	}

	return append(rows, "")
}

func printDescriptionAndUsage(cmd *cobra.Command) ([]string, bool) {
	// We need to manually add the -h flag so the usage line is suffixed with "[flags]".
	cmd.InitDefaultHelpFlag()

	rows := []string{
		printLongestDescription(cmd),
		"",
		"::",
		"",
		fmt.Sprintf("  %s", cmd.UseLine()),
		"",
	}
	return rows, true
}

func printNotes(cmd *cobra.Command, depth int) []string {
	var rows []string

	if strings.HasPrefix(cmd.CommandPath(), "confluent local") {
		include := strings.Repeat("../", depth) + "includes/path-set-cli.rst"
		rows = append(rows, printSphinxBlock("include", include, nil)...)
	}

	if strings.HasPrefix(cmd.CommandPath(), "confluent secret") {
		tip := "For examples, see :platform:`Secrets Usage Examples|security/secrets.html#secrets-examples`."
		rows = append(rows, printSphinxBlock("tip", tip, nil)...)
	}

	if cmd.CommandPath() == "confluent iam rbac role-binding create" {
		note := "If you need to troubleshoot when setting up role bindings, it may be helpful to view audit logs on the fly to identify authorization events for specific principals, resources, or operations. For details, refer to :platform:`Viewing audit logs on the fly|security/audit-logs/audit-logs-properties-config.html#view-audit-logs-on-the-fly`."
		rows = append(rows, printSphinxBlock("note", note, nil)...)
	}

	return rows
}

func printFlags(cmd *cobra.Command) ([]string, bool) {
	pcmd.LabelRequiredFlags(cmd)
	return printFlagSet(cmd.NonInheritedFlags())
}

func printGlobalFlags(cmd *cobra.Command) ([]string, bool) {
	pcmd.LabelRequiredFlags(cmd)
	return printFlagSet(cmd.InheritedFlags())
}

func printFlagSet(flags *pflag.FlagSet) ([]string, bool) {
	if !flags.HasAvailableFlags() {
		return []string{"No flags.", ""}, false
	}

	buf := new(bytes.Buffer)
	flags.SetOutput(buf)
	flags.PrintDefaults()

	rows := []string{
		"::",
		"",
		buf.String(),
	}
	return rows, true
}

func printExamples(cmd *cobra.Command) ([]string, bool) {
	if cmd.Example == "" {
		return []string{"No examples.", ""}, false
	}

	var rows []string
	isInsideCodeBlock := false

	for _, line := range strings.Split(cmd.Example, "\n") {
		if strings.HasPrefix(line, "  ") {
			// This line contains code. Write a "::" if this is start of a code block
			if !isInsideCodeBlock {
				rows = append(rows, "")
				rows = append(rows, "::")
				rows = append(rows, "")
			}

			// Strip the tab and shell prompt
			line = strings.TrimPrefix(line, "  ")
			line = strings.TrimPrefix(line, "$ ")
			rows = append(rows, "  "+line)

			isInsideCodeBlock = true
		} else if line != "" {
			rows = append(rows, "")
			rows = append(rows, formatReST(line))

			isInsideCodeBlock = false
		}
	}

	rows = rows[1:]
	rows = append(rows, "")

	return rows, true
}

func printSeeAlso(cmd *cobra.Command) []string {
	return []string{
		fmt.Sprintf("* %s - %s", printSphinxRef(cmd.Parent()), cmd.Parent().Short),
		"",
	}
}
