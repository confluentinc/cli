package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *mirrorCommand) newReverseAndPauseMirrorCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "reverse-and-pause <destination-topic-1> [destination-topic-2] ... [destination-topic-N]",
		Short:             "Reverses local mirror topics & pauses remote mirror topics.",
		RunE:              c.reverseAndPauseMirror,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Reverses local mirror topics and pauses remote mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror reverse-and-pause my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually reverse the local mirror topic and pause the remote mirror topic, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) reverseAndPauseMirror(cmd *cobra.Command, args []string) error {
	link, err := cmd.Flags().GetString("link")
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

	data := kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: &args}

	results, err := kafkaREST.CloudClient.UpdateKafkaMirrorTopicsReverseAndPauseMirror(link, dryRun, data)
	if err != nil {
		return err
	}

	return printAlterMirrorResult(cmd, results)
}
