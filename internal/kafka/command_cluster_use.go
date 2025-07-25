package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *clusterCommand) newUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Use a Kafka cluster in subsequent commands.",
		Long:              "Choose a Kafka cluster to be used in subsequent commands which support passing a cluster with the `--cluster` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	return cmd
}

func (c *clusterCommand) use(_ *cobra.Command, args []string) error {
	id := args[0]

	if _, err := kafka.FindCluster(c.V2Client, c.Context, id); err != nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, id),
			errors.ChooseRightEnvironmentSuggestions,
		)
	}

	c.Context.KafkaClusterContext.SetActiveKafkaCluster(id)
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.ErrPrintf(c.Config.EnableColor, "Set Kafka cluster \"%s\" as the active cluster for environment \"%s\".\n", id, c.Context.GetCurrentEnvironment())

	if activeEndpoint := c.Context.KafkaClusterContext.GetActiveKafkaClusterEndpoint(); activeEndpoint != "" {
		c.Context.KafkaClusterContext.UnsetActiveKafkaClusterEndpoint()
		output.ErrPrintf(c.Config.EnableColor, "Unset current endpoint. "+
			"Use `kafka cluster endpoint list` to view the available endpoints and use `cluster endpoint use` to set an active endpoint for the current Kafka cluster.\n")
	}

	return nil
}
