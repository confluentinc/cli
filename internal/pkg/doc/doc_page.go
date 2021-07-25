package doc

import (
	"bytes"
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

// generateDocPage creates a file which contains the command description, usage, flags, examples, and more.
// If there are multiple versions of a single command, tabs are used within the file.
func generateDocPage(tabs []Tab, dir string, depth int) error {
	name := strings.ReplaceAll(tabs[0].Command.CommandPath(), " ", "_")
	path := filepath.Join(dir, name+".rst")
	page := printTabbedDocPage(tabs, depth)
	return writeFile(path, page)
}

func printTabbedDocPage(tabs []Tab, depth int) string {
	pages := make([]string, len(tabs))
	for i, cmd := range tabs {
		pages[i] = printDocPage(cmd.Command, depth)
	}
	return printTabbedPages(tabs, pages)
}

func printDocPage(cmd *cobra.Command, depth int) string {
	rows := flatten([][]string{
		printHeader(cmd, "-"),
		printWarnings(cmd, depth),
		printDescription(cmd),
		printUsage(cmd),
		printNotes(cmd, depth),
		printFlags(cmd),
		printExamples(cmd),
		printSeeAlso(cmd),
	})

	return strings.Join(rows, "\n")
}

func flatten(arrs [][]string) []string {
	var flatArr []string
	for _, arr := range arrs {
		flatArr = append(flatArr, arr...)
	}
	return flatArr
}

func printHeader(cmd *cobra.Command, underline string) []string {
	name := cmd.CommandPath()

	return []string{
		fmt.Sprintf(".. _%s:", printRef(cmd)),
		"",
		name,
		strings.Repeat(underline, len(name)),
		"",
	}
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
	// TODO: Make val optional

	rows := []string{
		fmt.Sprintf(".. %s:: %s", key, val),
	}

	var keys []string
	for key := range args {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	for _, key := range keys {
		rows = append(rows, fmt.Sprintf("  :%s: %s", key, args[key]))
	}

	return append(rows, "")
}

func printDescription(cmd *cobra.Command) []string {
	return []string{
		"Description",
		"~~~~~~~~~~~",
		"",
		printLongestDescription(cmd),
		"",
	}
}

func printUsage(cmd *cobra.Command) []string {
	return []string{
		"::",
		"",
		fmt.Sprintf("  %s", cmd.UseLine()),
		"",
	}
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

	if cmd.CommandPath() == "confluent iam rolebinding create" {
		note := "If you need to troubleshoot when setting up role bindings, it may be helpful to view audit logs on the fly to identify authorization events for specific principals, resources, or operations. For details, refer to :platform:`Viewing audit logs on the fly|security/audit-logs/audit-logs-properties-config.html#view-audit-logs-on-the-fly`."
		rows = append(rows, printSphinxBlock("note", note, nil)...)
	}

	return rows
}

func printFlags(cmd *cobra.Command) []string {
	pcmd.LabelRequiredFlags(cmd)

	return flatten([][]string{
		printFlagSet("Flags", cmd.NonInheritedFlags()),
		printFlagSet("Global Flags", cmd.InheritedFlags()),
	})
}

func printFlagSet(title string, flags *pflag.FlagSet) []string {
	if !flags.HasAvailableFlags() {
		return []string{}
	}

	buf := new(bytes.Buffer)
	flags.SetOutput(buf)
	flags.PrintDefaults()

	return []string{
		title,
		strings.Repeat("~", len(title)),
		"",
		"::",
		"",
		buf.String(),
	}
}

func printExamples(cmd *cobra.Command) []string {
	if cmd.Example == "" {
		return []string{}
	}

	rows := []string{
		"Examples",
		"~~~~~~~~",
	}

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
			// This line contains text. Use double backticks for .rst code snippets.
			line = strings.ReplaceAll(line, "`", "``")
			rows = append(rows, "")
			rows = append(rows, line)

			isInsideCodeBlock = false
		}
	}

	return append(rows, "")
}

func printSeeAlso(cmd *cobra.Command) []string {
	var rows []string

	if cmd.HasParent() {
		rows = append(rows, fmt.Sprintf("* %s - %s", printSphinxRef(cmd.Parent()), cmd.Parent().Short))
	}

	for _, subcommand := range cmd.Commands() {
		if subcommand.IsAvailableCommand() {
			rows = append(rows, fmt.Sprintf("* %s - %s", printSphinxRef(subcommand), subcommand.Short))
		}
	}

	sort.Strings(rows)

	header := []string{
		"See Also",
		"~~~~~~~~",
		"",
	}
	rows = append(header, rows...)

	return append(rows, "")
}
