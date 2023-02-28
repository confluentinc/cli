package pipeline

import (
	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <pipeline-id>",
		Short: "Request to activate a pipeline.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.activate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Request to activate Stream Designer pipeline "pipe-12345".`,
				Code: `confluent pipeline activate pipe-12345`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) activate(cmd *cobra.Command, args []string) error {
	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	updatePipeline := streamdesignerv1.SdV1PipelineUpdate{Spec: &streamdesignerv1.SdV1PipelineSpecUpdate{Activated: streamdesignerv1.PtrBool(true)}}

	pipeline, err := c.V2Client.UpdateSdPipeline(c.EnvironmentId(cmd), cluster.ID, args[0], updatePipeline)
	if err != nil {
		return err
	}

	return print(cmd, pipeline)
}
