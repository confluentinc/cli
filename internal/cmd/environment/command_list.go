package environment

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/org"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name"}
	listHumanLabels      = []string{"ID", "Name"}
	listStructuredLabels = []string{"id", "name"}
)

type environmentStruct struct {
	Id   string
	Name string
}

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Confluent Cloud environments.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	resp, _, err := org.ListEnvironments(c.V2Client.OrgClient, c.AuthToken())
	if err != nil {
		return err
	}
	environments := resp.Data

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, environment := range environments {
		// Add '*' only in the case where we are printing out tables
		envStruct := environmentStruct{
			Id:   *environment.Id,
			Name: *environment.DisplayName,
		}
		if outputWriter.GetOutputFormat() == output.Human {
			if envStruct.Id == c.EnvironmentId() {
				envStruct.Id = fmt.Sprintf("* %s", envStruct.Id)
			} else {
				envStruct.Id = fmt.Sprintf("  %s", envStruct.Id)
			}
		}
		outputWriter.AddElement(&envStruct)
	}
	return outputWriter.Out()
}
