package kafka

import (
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
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

	if confirm, err := c.confirmDeletion(cmd, environmentId, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if r, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			return errors.CatchKafkaNotFoundError(err, id, r)
		}
		return nil
	}

	deletedIDs, err := resource.Delete(args, deleteFunc, resource.KafkaCluster)
	if err2 := c.removeKafkaClusterConfigs(deletedIDs); err2 != nil {
		err = multierror.Append(err, err2)
	}
	if err != nil {
		if len(args)-len(deletedIDs) > 1 {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the clusters are not associated with any active Connect clusters.")
		} else {
			return errors.NewErrorWithSuggestions(err.Error(), "Ensure the cluster is not associated with any active Connect clusters.")
		}
	}

	return nil
}

func (c *clusterCommand) confirmDeletion(cmd *cobra.Command, environmentId string, args []string) (bool, error) {
	if err := resource.ValidatePrefixes(resource.KafkaCluster, args); err != nil {
		return false, err
	}

	var displayName string
	describeFunc := func(id string) error {
		cluster, _, err := c.V2Client.DescribeKafkaCluster(id, environmentId)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = cluster.Spec.GetDisplayName()
		}

		return nil
	}

	err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.KafkaCluster, describeFunc)
	if err != nil {
		PluralClusterEnvironmentSuggestions := "Ensure the clusters you are specifying belong to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
		return false, errors.NewErrorWithSuggestions(err.Error(), PluralClusterEnvironmentSuggestions)
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.KafkaCluster, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.KafkaCluster, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}

func (c *clusterCommand) removeKafkaClusterConfigs(deletedIDs []string) error {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	for _, id := range deletedIDs {
		if err := c.Context.RemoveKafkaClusterConfig(id); err != nil {
			errs = multierror.Append(errs, err)
		}
	}

	return errs.ErrorOrNil()
}
