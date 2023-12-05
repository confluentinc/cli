package pipeline

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/kafka"
)

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <pipeline-id>",
		Short:             "Describe a Stream Designer pipeline.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline describe pipe-12345`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	cluster, err := kafka.GetClusterForCommand(c.V2Client, c.Context)
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	pipeline, err := c.V2Client.GetSdPipeline(environmentId, cluster.ID, args[0])
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
