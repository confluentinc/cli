package kafka

import (
	"github.com/antihax/optional"
	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *mirrorCommand) newFailoverCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "failover <destination-topic-1> <destination-topic-2> ... <destination-topic-N> --link my-link",
		Short: "Failover mirror topics.",
		Args:  cobra.MinimumNArgs(1),
		RunE:  c.failover,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Failover mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror failover my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired(linkFlagName)

	return cmd
}

func (c *mirrorCommand) failover(cmd *cobra.Command, args []string) error {
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

	failoverMirrorOpt := &kafkarestv3.ClustersClusterIdLinksLinkNameMirrorsfailoverPostOpts{
		AlterMirrorsRequestData: optional.NewInterface(kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: args}),
		ValidateOnly:            optional.NewBool(validateOnly),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingApi.ClustersClusterIdLinksLinkNameMirrorsfailoverPost(kafkaREST.Context, lkc, linkName, failoverMirrorOpt)
	if err != nil {
		return kafkaRestError(kafkaREST.Client.GetConfig().BasePath, err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}
