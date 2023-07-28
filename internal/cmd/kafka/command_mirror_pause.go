package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

func (c *mirrorCommand) newPauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "pause <destination-topic-1> [destination-topic-2] ... [destination-topic-N]",
		Short:             "Pause mirror topics.",
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		RunE:              c.pause,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Pause mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror pause my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired(linkFlagName))

	return cmd
}

func (c *mirrorCommand) pause(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if err != nil {
		return err
	}

	cluster, err := c.Context.GetKafkaClusterForCommand()
	if err != nil {
		return err
	}

	pauseMirrorOpt := &kafkarestv3.UpdateKafkaMirrorTopicsPauseOpts{
		AlterMirrorsRequestData: optional.NewInterface(kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: args}),
		ValidateOnly:            optional.NewBool(dryRun),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsPause(kafkaREST.Context, cluster.ID, linkName, pauseMirrorOpt)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}
