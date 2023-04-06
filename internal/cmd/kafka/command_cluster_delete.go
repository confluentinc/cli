package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/deletion"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *clusterCommand) newDeleteCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <id-1> [id-2] ... [id-n]",
		Short:             "Delete Kafka clusters.",
		Args:              cobra.MinimumNArgs(1),
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
	environmentId, err := c.EnvironmentId()
	if err != nil {
		return err
	}

	displayName, err := c.validateArgs(cmd, environmentId, args)
	if err != nil {
		return err
	}

	if len(args) == 1 {
		if err := form.ConfirmDeletionWithString(cmd, resource.KafkaCluster, args[0], displayName); err != nil {
			return err
		}
	} else {
		if ok, err := form.ConfirmDeletionYesNo(cmd, resource.KafkaCluster, args); err != nil || !ok {
			return err
		}
	}

	var errs error
	var deleted []string
	for _, id := range args {
		if r, err := c.V2Client.DeleteKafkaCluster(id, environmentId); err != nil {
			errs = errors.Join(errs, errors.CatchKafkaNotFoundError(err, id, r))
		} else {
			deleted = append(deleted, id)
			if err := c.Context.RemoveKafkaClusterConfig(id); err != nil {
				errs = errors.Join(errs, err)
			}
		}
	}
	deletion.PrintSuccessfulDeletionMsg(deleted, resource.KafkaCluster)

	if errs != nil {
		if len(args)-len(deleted) > 1 {
			return errors.NewErrorWithSuggestions(errs.Error(), "Ensure the clusters are not associated with any active Connect clusters.")
		} else {
			return errors.NewErrorWithSuggestions(errs.Error(), "Ensure the cluster is not associated with any active Connect clusters.")
		}
	}

	return nil
}

func (c *clusterCommand) validateArgs(cmd *cobra.Command, environmentId string, args []string) (string, error) {
	if err := resource.ValidatePrefixes(resource.KafkaCluster, args); err != nil {
		return "", err
	}

	var displayName string
	describeFunc := func(id string) error {
		cluster, _, err := c.V2Client.DescribeKafkaCluster(id, environmentId)
		if err == nil && displayName == "" { // store the first valid cluster name
			displayName = cluster.Spec.GetDisplayName()
		}
		return err
	}

	err := deletion.ValidateArgsForDeletion(cmd, args, resource.KafkaCluster, describeFunc)
	if err != nil {
		PluralClusterEnvironmentSuggestions := "Ensure the clusters you are specifying belong to the currently selected environment with `confluent kafka cluster list`, `confluent environment list`, and `confluent environment use`."
		err = errors.NewErrorWithSuggestions(err.Error(), PluralClusterEnvironmentSuggestions)
	}

	return displayName, err
}
