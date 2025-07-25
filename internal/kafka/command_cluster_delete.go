package kafka

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/deletion"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete one or more Kafka clusters.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Annotations:       map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.DescribeKafkaCluster(id, environmentId)
		return err == nil
	}

	if err := deletion.ValidateAndConfirm(cmd, args, existenceFunc, resource.KafkaCluster); err != nil {
		PluralClusterEnvironmentSuggestions := "Ensure the clusters you are specifying belong to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
		return errors.NewErrorWithSuggestions(err.Error(), PluralClusterEnvironmentSuggestions)
	}

	deleteFunc := func(id string) error {
		if httpResp, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			return errors.CatchKafkaNotFoundError(err, id, httpResp)
		}
		return nil
	}

	deletedIds, err := deletion.Delete(cmd, args, deleteFunc, resource.KafkaCluster)

	errs := multierror.Append(err, c.removeKafkaClusterConfigs(deletedIds))
	if errs.ErrorOrNil() != nil {
		if len(args)-len(deletedIds) > 1 {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the clusters are not associated with any active Connect clusters.")
		} else {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the cluster is not associated with any active Connect clusters.")
		}
	}

	return nil
}

func (c *clusterCommand) removeKafkaClusterConfigs(deletedIds []string) error {
	for _, id := range deletedIds {
		c.Context.KafkaClusterContext.RemoveKafkaCluster(id)
	}

	return c.Context.Save()
}
