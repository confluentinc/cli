package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newDeleteCommand(prerunner pcmd.PreRunner) *cobra.Command {
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

	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) delete(cmd *cobra.Command, args []string) error {
	// get kafka cluster
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	// call api
	err = c.V2Client.DeleteSdPipeline(c.EnvironmentId(), cluster.ID, args[0])
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.DeletedResourceMsg, resource.Pipeline, args[0])
	return nil
}
