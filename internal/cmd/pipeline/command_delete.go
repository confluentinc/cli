package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "delete <pipeline-id-1> [pipeline-id-2] ... [pipeline-id-n]",
		Short:             "Delete one or more pipelines.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.delete,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to delete Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline delete pipe-12345`,
			},
		),
	}

	pcmd.AddForceFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	if confirm, err := c.confirmDeletion(cmd, environmentId, cluster.ID, args); err != nil {
		return err
	} else if !confirm {
		return nil
	}

	deleteFunc := func(id string) error {
		if err := c.V2Client.DeleteSdPipeline(environmentId, cluster.ID, id); err != nil {
			return err
		}
		return nil
	}

	deleted, err := resource.Delete(args, deleteFunc, nil)
	if len(deleted) == 1 {
		output.Printf("Requested to delete pipeline \"%s\".\n", deleted[0])
	} else if len(deleted) > 1 {
		output.Printf("Requested to delete pipelines %s.\n", utils.ArrayToCommaDelimitedString(deleted, "and"))
	}

	return err
}

func (c *command) confirmDeletion(cmd *cobra.Command, environmentId, clusterId string, args []string) (bool, error) {
	var displayName string
	describeFunc := func(id string) error {
		pipeline, err := c.V2Client.GetSdPipeline(environmentId, clusterId, id)
		if err != nil {
			return err
		}
		if id == args[0] {
			displayName = pipeline.Spec.GetDisplayName()
		}

		return nil
	}

	if err := resource.ValidateArgs(pcmd.FullParentName(cmd), args, resource.Pipeline, describeFunc); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return form.ConfirmDeletionYesNo(cmd, form.DefaultYesNoPromptString(resource.Pipeline, args))
	}

	if err := form.ConfirmDeletionWithString(cmd, form.DefaultPromptString(resource.Pipeline, args[0], displayName), displayName); err != nil {
		return false, err
	}

	return true, nil
}
