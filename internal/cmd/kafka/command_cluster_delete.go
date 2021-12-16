package kafka

import (
	"context"

	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/analytics"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "delete <id>",
		Short:       "Delete a Kafka cluster.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.delete),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	req := &schedv1.KafkaCluster{AccountId: c.EnvironmentId(), Id: args[0]}
	err := c.Client.Kafka.Delete(context.Background(), req)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}

	if err := c.Context.RemoveKafkaClusterConfig(args[0]); err != nil {
		return err
	}

	utils.Printf(cmd, errors.KafkaClusterDeletedMsg, args[0])
	c.analyticsClient.SetSpecialProperty(analytics.ResourceIDPropertiesKey, args[0])
	return nil
}
