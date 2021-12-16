package kafka

import (
	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *mirrorCommand) newPauseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pause <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link my-link",
		Short: "Pause mirror topics.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.pause,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Pause mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror pause my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(linkFlagName)

	return cmd
}
func (c *mirrorCommand) pause(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	validateOnly, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST()
	if kafkaREST == nil {
		if err != nil {
			return err
		}
		return errors.New(errors.RestProxyNotAvailableMsg)
	}

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedStateFlagCommand)
	if err != nil {
		return err
	}

	pauseMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorspausePostOpts{
		AlterMirrorsRequestData: optional.NewInterface(kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: args}),
		ValidateOnly:            optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameMirrorspausePost(kafkaREST.Context, lkc, linkName, pauseMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}
