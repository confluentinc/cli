package environment

import (
	"context"
	"fmt"

	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	listFields           = []string{"Id", "Name"}
	listHumanLabels      = []string{"ID", "Name"}
	listStructuredLabels = []string{"id", "name"}
)

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
	environments, err := c.Client.Account.List(context.Background(), &orgv1.Account{})
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, listFields, listHumanLabels, listStructuredLabels)
	if err != nil {
		return err
	}
	for _, environment := range environments {
		// Add '*' only in the case where we are printing out tables
		if outputWriter.GetOutputFormat() == output.Human {
			if environment.Id == c.EnvironmentId() {
				environment.Id = fmt.Sprintf("* %s", environment.Id)
			} else {
				environment.Id = fmt.Sprintf("  %s", environment.Id)
			}
		}
		outputWriter.AddElement(environment)
	}
	return outputWriter.Out()
}
