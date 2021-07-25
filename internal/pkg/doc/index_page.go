package doc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

// generateIndexPage creates a file called index.rst which contains the command description and links to subcommands.
// If there are multiple versions of a single command, tabs are used within index.rst.
func generateIndexPage(tabs []Tab, dir string, printIndexHeader func(*cobra.Command) []string) error {
	name := tabs[0].Command.Name()
	path := filepath.Join(dir, name, "index.rst")
	page := printTabbedIndexPage(tabs, printIndexHeader)
	return writeFile(path, page)
}

func printTabbedIndexPage(tabs []Tab, printIndexHeader func(*cobra.Command) []string) string {
	pages := make([]string, len(tabs))
	for i, cmd := range tabs {
		pages[i] = printIndexPage(cmd.Command, printIndexHeader)
	}
	return printTabbedPages(tabs, pages)
}

func printIndexPage(cmd *cobra.Command, printIndexHeader func(*cobra.Command) []string) string {
	rows := flatten([][]string{
		printIndexHeader(cmd),
		printTableOfContents(cmd),
		printSubcommands(cmd),
	})
	return strings.Join(rows, "\n")
}

func printIndexHeader(cmd *cobra.Command) []string {
	return flatten([][]string{
		printHeader(cmd, "="),
		printDescription(cmd),
	})
}

func printLongestDescription(cmd *cobra.Command) string {
	description := cmd.Short
	if cmd.Long != "" {
		description = cmd.Long
	}

	// ReST uses double backticks for code snippets.
	return strings.ReplaceAll(description, "`", "``")
}

func printRootIndexHeader(_ *cobra.Command) []string {
	return []string{
		".. _confluent-ref:",
		"",
		"|confluent| CLI Command Reference",
		"=================================",
		"",
		"The available |confluent| CLI commands are documented here.",
		"",
	}
}

func printTableOfContents(cmd *cobra.Command) []string {
	rows := []string{
		".. toctree::",
		"   :hidden:",
		"",
	}

	for _, subcommand := range cmd.Commands() {
		if subcommand.IsAvailableCommand() {
			link := printLink(subcommand)
			rows = append(rows, fmt.Sprintf("   %s", link))
		}
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

func printSubcommands(cmd *cobra.Command) []string {
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
	return dedent(rows)
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
