package doc

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/lithammer/dedent"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func GenReSTIndex(cmd *cobra.Command, filename string, filePrepender func(*cobra.Command) string, linkHandler func(string) string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	commands := genReSTIndex(cmd)

	// Title
	if _, err := io.WriteString(f, filePrepender(cmd)); err != nil {
		return err
	}

	// Navigation
	fmt.Fprintf(f, ".. toctree::\n   :hidden:\n\n")
	for _, c := range commands {
		fmt.Fprintf(f, "   %s\n", c.ref)
	}
	fmt.Fprintln(f)

	// Write to a buffer so we can dedent before we print.
	//
	// This is needed because a space for center separator between columns also creates a space on the left,
	// effectively indenting the table by a space. This messes up ReST which views that as a blockquote.
	buf := new(bytes.Buffer)

	table := tablewriter.NewWriter(buf)
	table.SetAutoWrapText(false)
	table.SetColumnSeparator(" ")
	table.SetCenterSeparator(" ")
	table.SetRowSeparator("=")
	table.SetAutoFormatHeaders(false)

	table.SetHeader([]string{"Command", "Description"})
	for _, c := range commands {
		row := []string{linkHandler(c.command), c.description}
		table.Append(row)
	}
	table.Render()

	_, err = io.WriteString(f, dedent.Dedent(buf.String()))
	return err
}

type command struct {
	command     string
	ref         string
	description string
}

func genReSTIndex(cmd *cobra.Command) []command {
	var allCommands []command

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}

		name, ref := link(c)
		child := command{command: name, ref: ref, description: c.Short}
		allCommands = append(allCommands, child)
	}

	return allCommands
}

func link(cmd *cobra.Command) (string, string) {
	path := strings.ReplaceAll(fullCommand(cmd), " ", "_")

	ref := path
	if cmd.HasSubCommands() {
		x := strings.Split(path, "_")
		ref = filepath.Join(x[len(x)-1], "index")
	}

	return path, ref
}

func fullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Name()}
	cmd.VisitParents(func(command *cobra.Command) {
		use = append([]string{command.Use}, use...)
	})
	return strings.Join(use, " ")
}

// The header for all indexes other than the root.
func indexHeader(command *cobra.Command) string {
	buf := new(bytes.Buffer)

	name := command.CommandPath()
	buf.WriteString(fmt.Sprintf(".. _%s:\n\n", strings.ReplaceAll(name, " ", "_")))
	buf.WriteString(fmt.Sprintf("%s\n", name))
	buf.WriteString(strings.Repeat("=", len(name)) + "\n\n")

	// Description
	desc := command.Short
	if command.Long != "" {
		desc = command.Long
	}
	buf.WriteString("Description\n")
	buf.WriteString("~~~~~~~~~~~\n\n")
	buf.WriteString(desc + "\n\n")

	return buf.String()
}
