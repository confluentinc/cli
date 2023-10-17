package pipeline

import (
	"github.com/spf13/cobra"

	streamdesignerv1 "github.com/confluentinc/ccloud-sdk-go-v2/stream-designer/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "activate <pipeline-id>",
		Short:             "Request to activate a pipeline.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.activate,
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
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand(c.V2Client)
	if err != nil {
		return err
	}

	pipeline := streamdesignerv1.SdV1Pipeline{Spec: &streamdesignerv1.SdV1PipelineSpec{
		Activated:    streamdesignerv1.PtrBool(true),
		Environment:  &streamdesignerv1.ObjectReference{Id: environmentId},
		KafkaCluster: &streamdesignerv1.ObjectReference{Id: cluster.ID},
	}}

	pipeline, err = c.V2Client.UpdateSdPipeline(args[0], pipeline)
	if err != nil {
		return err
	}

	return printTable(cmd, pipeline)
}
