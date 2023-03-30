package docs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/types"
)

const tab = "   "

// generateIndexPage creates a file called index.rst which contains the command description and links to subcommands.
// If there are multiple versions of a single command, tabs are used within index.rst.
func generateIndexPage(tabs []Tab, dir string) error {
	if cmd := tabs[0].Command; cmd == cmd.Root() {
		rows := printIndexPage(tabs, true)
		if err := writeFile(filepath.Join(dir, "overview.rst"), strings.Join(rows, "\n")); err != nil {
			return err
		}

		rows = printRootIndexPage(tabs)
		return writeFile(filepath.Join(dir, "index.rst"), strings.Join(rows, "\n"))
	}

	rows := printIndexPage(tabs, false)
	return writeFile(filepath.Join(dir, "index.rst"), strings.Join(rows, "\n"))
}

func printRootIndexPage(tabs []Tab) []string {
	cmd := tabs[0].Command

	return flatten([][]string{
		printComments(),
		printHeader(cmd, false),
		printTitle(cmd, "="),
		printInlineScript(),
		printTableOfContents(tabs),
	})
}

func printIndexPage(tabs []Tab, isOverview bool) []string {
	cmd := tabs[0].Command

	rows := [][]string{
		printComments(),
		printHeader(cmd, isOverview),
		printTitle(cmd, "="),
		printSection("Aliases", printAliases(cmd)),
		printTabbedSection("Description", printDescription, tabs),
	}

	if cmd := tabs[0].Command; cmd != cmd.Root() {
		rows = append(rows, printTableOfContents(tabs))
	}

	rows = append(rows, printTabbedSection("Subcommands", printSubcommands, tabs))

	return flatten(rows)
}

func flatten(arrs [][]string) []string {
	var flatArr []string
	for _, arr := range arrs {
		flatArr = append(flatArr, arr...)
	}
	return flatArr
}

func printComments() []string {
	return []string{
		"..",
		tab + "WARNING: This documentation is auto-generated from the confluentinc/cli repository and should not be manually edited.",
		"",
	}
}

func printHeader(cmd *cobra.Command, isOverview bool) []string {
	return []string{
		fmt.Sprintf(".. _%s:", printRef(cmd, isOverview)),
		"",
	}
}

func printTitle(cmd *cobra.Command, underline string) []string {
	name := cmd.CommandPath()
	if cmd == cmd.Root() {
		name = "|confluent| CLI Command Reference"
	}

	return []string{
		name,
		strings.Repeat(underline, len(name)),
		"",
	}
}

func printInlineScript() []string {
	return []string{
		".. raw:: html",
		"",
		tab + `<script type="text/javascript">`,
		tab + tab + "window.location = 'overview.html';",
		tab + "</script>",
		"",
	}
}

func printTableOfContents(tabs []Tab) []string {
	// Combine subcommands across tabs, removing duplicates
	linksByName := make(map[string]string)
	for _, tab := range tabs {
		for _, subcommand := range tab.Command.Commands() {
			if subcommand.IsAvailableCommand() {
				linksByName[subcommand.Name()] = printLink(subcommand)
			}
		}
	}

	names := types.GetSortedKeys(linksByName)

	rows := []string{
		".. toctree::",
	}

	if cmd := tabs[0].Command; cmd == cmd.Root() {
		rows = append(rows, tab+":maxdepth: 1")
	}

	rows = append(rows, tab+":hidden:", "")

	if cmd := tabs[0].Command; cmd == cmd.Root() {
		rows = append(rows, tab+"Overview <overview>")
	}

	for _, name := range names {
		rows = append(rows, tab+linksByName[name])
	}

	return append(rows, "")
}

func printLink(cmd *cobra.Command) string {
	if cmd.HasSubCommands() {
		// Example: command/index
		return path.Join(cmd.Name(), "index")
	} else {
		return printRef(cmd, false)
	}
}

func printAliases(cmd *cobra.Command) []string {
	if len(cmd.Aliases) == 0 {
		return []string{}
	}

	aliases := append([]string{cmd.Name()}, cmd.Aliases...)

	return []string{
		"::",
		"",
		fmt.Sprintf("  %s", strings.Join(aliases, ", ")),
		"",
	}
}

func printDescription(cmd *cobra.Command) ([]string, bool) {
	var rows []string

	if cmd == cmd.Root() {
		rows = []string{
			"The available |confluent| CLI commands are documented here.",
			"",
		}
	} else {
		rows = []string{
			printLongestDescription(cmd),
			"",
		}
	}

	return rows, true
}

func printLongestDescription(cmd *cobra.Command) string {
	description := cmd.Short
	if cmd.Long != "" {
		description = cmd.Long
	}

	return formatReST(description)
}

func formatReST(s string) string {
	// ReST uses double backticks for code snippets.
	s = strings.ReplaceAll(s, "`", "``")

	// ReST targets are formatted like "target_" and can be added to a string inline.
	// We escape the underscore because none of our CLI descriptions or examples include ReST targets.
	matches := regexp.MustCompile(`[0-9A-Za-z]+(_)[^0-9A-Za-z]`).FindAllStringSubmatchIndex(s, -1)
	for i := len(matches) - 1; i >= 0; i-- {
		lo := matches[i][2]
		hi := matches[i][3]
		s = fmt.Sprintf(`%s\_%s`, s[:lo], s[hi:])
	}

	return s
}

func printSubcommands(cmd *cobra.Command) ([]string, bool) {
	buf := new(bytes.Buffer)

	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")
	table.SetRowSeparator("=")
	table.SetAutoFormatHeaders(false)

	table.SetHeader([]string{"Command", "Description"})
	for _, subcommand := range cmd.Commands() {
		if subcommand.IsAvailableCommand() {
			table.Append([]string{printSphinxRef(subcommand), subcommand.Short})
		}
	}
	table.Render()

	rows := strings.Split(buf.String(), "\n")

	// The tablewriter library leaves a leading and trailing character of whitespace on every row.
	// Remove them to conform to ReST syntax.
	return dedent(rows), true
}

func printSphinxRef(cmd *cobra.Command) string {
	ref := printRef(cmd, false)
	return fmt.Sprintf(":ref:`%s`", ref)
}

func printRef(cmd *cobra.Command, isOverview bool) string {
	// Example: command_subcommand
	ref := strings.ReplaceAll(cmd.CommandPath(), " ", "_")

	// The root ref is named "confluent-ref"
	if cmd == cmd.Root() {
		ref += "-ref"
		if isOverview {
			ref += "-index"
		}
	}

	return ref
}

func dedent(rows []string) []string {
	for i := range rows {
		rows[i] = strings.TrimPrefix(rows[i], " ")
		rows[i] = strings.TrimSuffix(rows[i], " ")
	}
	return rows
}

func writeFile(path, text string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = io.WriteString(file, text)
	return err
}
