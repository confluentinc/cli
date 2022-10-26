package environment

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name"}
	listHumanLabels      = []string{"ID", "Name"}
	listStructuredLabels = []string{"id", "name"}
)

type environmentRow struct {
	Id   string
	Name string
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud environments.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environments, err := c.V2Client.ListOrgEnvironments()
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, env := range environments {
		// Add '*' only in the case where we are printing out tables
		envRow := &environmentRow{
			Id:   *env.Id,
			Name: *env.DisplayName,
		}
		if outputWriter.GetOutputFormat() == output.Human {
			if envRow.Id == c.EnvironmentId() {
				envRow.Id = fmt.Sprintf("* %s", envRow.Id)
			} else {
				envRow.Id = fmt.Sprintf("  %s", envRow.Id)
			}
		}
		outputWriter.AddElement(envRow)
	}
	return outputWriter.Out()
}
