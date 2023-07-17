package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *clusterCommand) newUseCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id|name>",
		Short:             "Use a Kafka cluster in subsequent commands.",
		Long:              "Choose a Kafka cluster to be used in subsequent commands which support passing a cluster with the `--cluster` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.use,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterId := args[0]

	if _, err := c.Context.FindKafkaCluster(clusterId); err != nil {
		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return err
		}
		if clusterId, environmentId, err = c.clusterAndEnvNamesToIds(clusterId, environmentId); err != nil {
			return err
		}
		c.Context.SetCurrentEnvironment(environmentId)
		if _, err := c.Context.FindKafkaCluster(clusterId); err != nil {
			return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterId), errors.ChooseRightEnvironmentSuggestions)
		}
	}

	if err := c.Context.SetActiveKafkaCluster(clusterId); err != nil {
		return err
	}

	output.ErrPrintf(errors.UseKafkaClusterMsg, clusterId, c.Context.GetCurrentEnvironment())
	return nil
}
