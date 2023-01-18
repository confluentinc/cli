package pipeline

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <pipeline-id>",
		Short: "Delete a pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.delete,
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

	pipeline, err := c.V2Client.GetSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	promptMsg := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resource.Pipeline, pipeline.GetId(), pipeline.Spec.GetDisplayName())
	if _, err := form.ConfirmDeletion(cmd, promptMsg, pipeline.Spec.GetDisplayName()); err != nil {
		return err
	}

	if err := c.V2Client.DeleteSdPipeline(c.EnvironmentId(), cluster.ID, args[0]); err != nil {
		return err
	}

	utils.Printf(cmd, errors.RequestedDeleteResourceMsg, resource.Pipeline, args[0])
	return nil
}
