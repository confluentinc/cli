package doc

import (
	"io"
	"os"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func GenReSTIndex(cmd *cobra.Command, filename string, filePrepender func(string) string, linkHandler func(string, string) string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	commands, err := genReSTIndex(cmd, linkHandler)
	if err != nil {
		return err
	}

	if _, err := io.WriteString(f, filePrepender(filename)); err != nil {
		return err
	}

	table := tablewriter.NewWriter(f)
	table.SetAutoWrapText(false)
	table.SetColumnSeparator(" ")

	table.SetHeader([]string{"Command", "Description"})
	for _, c := range commands {
		row := []string{c.command, c.description}
		table.Append(row)
	}
	table.Render()

	return nil
}

type command struct {
	command string
	description string
}

func genReSTIndex(cmd *cobra.Command, linkHandler func(string, string) string) ([]command, error) {
	cname := fullCommand(cmd)
	ref := strings.Replace(cname, " ", "_", -1)
	allCommands := []command{{command: linkHandler(cname, ref), description: cmd.Short}}

	for _, c := range cmd.Commands() {
		if !c.IsAvailableCommand() || c.IsAdditionalHelpTopicCommand() {
			continue
		}
		commands, err := genReSTIndex(c, linkHandler)
		if err != nil {
			return nil, err
		}
		allCommands = append(allCommands, commands...)
	}
	return allCommands, nil
}

func fullCommand(cmd *cobra.Command) string {
	use := []string{cmd.Use}
	cmd.VisitParents(func(command *cobra.Command) {
		use = append([]string{command.Use}, use...)
	})
	return strings.Join(use, " ")
}
