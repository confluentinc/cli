package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/cmk"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Kafka cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              pcmd.NewCLIRunE(c.delete),
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	_, err := cmk.DeleteKafkaCluster(c.CmkClient, args[0], c.EnvironmentId(), c.AuthToken())
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0])
	}

	if err := c.Context.RemoveKafkaClusterConfig(args[0]); err != nil {
		return err
	}

	utils.Printf(cmd, errors.KafkaClusterDeletedMsg, args[0])
	return nil
}
