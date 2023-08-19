package kafka

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/deletion"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand(cfg *config.Config) *cobra.Command {
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
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) delete(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	PluralClusterEnvironmentSuggestions := "Ensure the clusters you are specifying belong to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
	cluster, _, err := c.V2Client.DescribeKafkaCluster(args[0], environmentId)
	if err != nil {
		return errors.NewErrorWithSuggestions(resource.ResourcesNotFoundError(cmd, resource.KafkaCluster, args[0]).Error(), PluralClusterEnvironmentSuggestions)
	}

	existenceFunc := func(id string) bool {
		_, _, err := c.V2Client.DescribeKafkaCluster(id, environmentId)
		return err == nil
	}

	if confirm, err := deletion.ValidateAndConfirmDeletion(cmd, args, existenceFunc, resource.KafkaCluster, cluster.Spec.GetDisplayName()); err != nil {
		return errors.NewErrorWithSuggestions(err.Error(), PluralClusterEnvironmentSuggestions)
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			return errors.CatchKafkaNotFoundError(err, id, r)
		}
		return nil
	}

	deletedIDs, err := deletion.Delete(args, deleteFunc, resource.KafkaCluster)

	errs := multierror.Append(err, c.removeKafkaClusterConfigs(deletedIDs))
	if errs.ErrorOrNil() != nil {
		if len(args)-len(deletedIDs) > 1 {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the clusters are not associated with any active Connect clusters.")
		} else {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the cluster is not associated with any active Connect clusters.")
		}
	}

	return nil
}

func (c *clusterCommand) removeKafkaClusterConfigs(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		errs = multierror.Append(errs, c.Context.RemoveKafkaClusterConfig(id))
	}

	return errs.ErrorOrNil()
}
