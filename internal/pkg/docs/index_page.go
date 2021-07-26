package docs

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// generateIndexPage creates a file called index.rst which contains the command description and links to subcommands.
// If there are multiple versions of a single command, tabs are used within index.rst.
func generateIndexPage(tabs []Tab, dir string) error {
	path := filepath.Join(dir, "index.rst")
	rows := printIndexPage(tabs)

	return writeFile(path, strings.Join(rows, "\n"))
}

func printIndexPage(tabs []Tab) []string {
	cmd := tabs[0].Command

	return flatten([][]string{
		printHeader(cmd),
		printTitle(cmd, "="),
		printTableOfContents(tabs),
		printTabbedSection("Description", printDescription, tabs),
		printTabbedSection("Subcommands", printSubcommands, tabs),
	})
}

func flatten(arrs [][]string) []string {
	var flatArr []string
	for _, arr := range arrs {
		flatArr = append(flatArr, arr...)
	}
	return flatArr
}

func printHeader(cmd *cobra.Command) []string {
	return []string{
		fmt.Sprintf(".. _%s:", printRef(cmd)),
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

	// Sort names
	var names []string
	for name := range linksByName {
		names = append(names, name)
	}
	sort.Strings(names)

	rows := []string{
		".. toctree::",
		"   :hidden:",
		"",
	}

	for _, name := range names {
		rows = append(rows, fmt.Sprintf("   %s", linksByName[name]))
	}

	return append(rows, "")
}

func printLink(cmd *cobra.Command) string {
	if cmd.HasSubCommands() {
		// Example: command/index
		x := strings.Split(cmd.CommandPath(), " ")
		return filepath.Join(x[len(x)-1], "index")
	} else {
		return printRef(cmd)
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
	return strings.ReplaceAll(s, "`", "``")
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
	ref := printRef(cmd)
	return fmt.Sprintf(":ref:`%s`", ref)
}

func printRef(cmd *cobra.Command) string {
	// Example: command_subcommand
	ref := strings.ReplaceAll(cmd.CommandPath(), " ", "_")

	// The root ref is named "confluent-ref"
	if cmd == cmd.Root() {
		ref += "-ref"
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
