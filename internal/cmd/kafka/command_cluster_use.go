package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newUseCommand() *cobra.Command {
	return &cobra.Command{
		Use:         "use <id>",
		Short:       "Make the Kafka cluster active for use in other commands.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.use),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}
}

func (c *clusterCommand) use(cmd *cobra.Command, args []string) error {
	clusterID := args[0]

	if _, err := c.Context.FindKafkaCluster(clusterID); err != nil {
		return errors.NewErrorWithSuggestions(fmt.Sprintf(errors.KafkaClusterNotFoundErrorMsg, clusterID), errors.ChooseRightEnvironmentSuggestions)
	}

	if err := c.Context.SetActiveKafkaCluster(clusterID); err != nil {
		return err
	}

	utils.ErrPrintf(cmd, errors.UseKafkaClusterMsg, clusterID, c.Context.GetCurrentEnvironmentId())
	return nil
}
