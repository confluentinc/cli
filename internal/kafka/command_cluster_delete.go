package kafka

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id>",
		Short:             "Delete a Kafka cluster.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.delete,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.FindKafkaCluster(args[0])
	if err != nil {
		// Replace the suggestions w/ the suggestions specific to delete requests
		return errors.NewErrorWithSuggestions(err.Error(), errors.KafkaClusterDeletingSuggestions)
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.KafkaCluster, args[0], cluster.GetName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, cluster.GetName()); err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	httpResp, err := c.V2Client.DeleteKafkaCluster(args[0], environmentId)
	if err != nil {
		return errors.CatchKafkaNotFoundError(err, args[0], httpResp)
	}

	if err := c.Context.RemoveKafkaClusterConfig(args[0]); err != nil {
		return err
	}

	output.Printf(errors.DeletedResourceMsg, resource.KafkaCluster, args[0])
	return nil
}
