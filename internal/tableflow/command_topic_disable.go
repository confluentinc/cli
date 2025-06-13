package tableflow

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newTopicDisableCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "disable <name-1> [name-2] ... [name-n]",
		Aliases:           []string{"delete"},
		Short:             "Disable topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validTopicArgsMultiple),
		RunE:              c.disable,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Disable a Tableflow topic "my-tableflow-topic" related to a Kafka cluster "lkc-123456".`,
				Code: "confluent tableflow topic disable my-tableflow-topic --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) disable(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, err := c.V2Client.GetTableflowTopic(environmentId, cluster.GetId(), id)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.Topic); err != nil {
		return err
	}

	deleteFunc := func(id string) error {
		return c.V2Client.DeleteTableflowTopic(environmentId, cluster.GetId(), id)
	}

	deletedTopic, err := deletion.DeleteWithoutMessage(cmd, args, deleteFunc)

	deleteMsg := "Requested to %s %s %s.\n"
	operation := cmd.CalledAs() // disable or delete depending on whether the user entered the alias
	if len(deletedTopic) == 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, operation, resource.Topic, fmt.Sprintf(`"%s"`, deletedTopic[0]))
	} else if len(deletedTopic) > 1 {
		output.Printf(c.Config.EnableColor, deleteMsg, operation, plural.Plural(resource.Topic), utils.ArrayToCommaDelimitedString(deletedTopic, "and"))
	}

	return err
}
