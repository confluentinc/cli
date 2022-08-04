package connect

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newResumeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "resume <id>",
		Short:             "Resume a connector.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.resume,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Resume a connector in the current or specified Kafka cluster context.",
				Code: "confluent connect resume",
			},
			examples.Example{
				Code: "confluent connect resume --cluster lkc-123456",
			},
		),
	}

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) resume(cmd *cobra.Command, args []string) error {
	kafkaCluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	connectorExpansion, err := c.V2Client.GetConnectorExpansionById(args[0], c.EnvironmentId(), kafkaCluster.ID)
	if err != nil {
		return err
	}

	if _, err := c.V2Client.ResumeConnector(connectorExpansion.Info.GetName(), c.EnvironmentId(), kafkaCluster.ID); err != nil {
		return err
	}

	utils.Printf(cmd, errors.ResumedConnectorMsg, args[0])
	return nil
}
