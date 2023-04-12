package kafka

import (
	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/kafka-rest-sdk-go/kafkarestv3"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/kafkarest"
)

func (c *mirrorCommand) newPromoteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "promote <destination-topic-1> [destination-topic-2] ... [destination-topic-N] --link my-link",
		Short: "Promote mirror topics.",
		RunE:  c.promote,
		Args:  cobra.MinimumNArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Promote mirror topics "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror promote my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	cmd.Flags().String(linkFlagName, "", "The name of the cluster link.")
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually create the link, but simply validates it.")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired(linkFlagName))

	return cmd
}

func (c *mirrorCommand) promote(cmd *cobra.Command, args []string) error {
	linkName, err := cmd.Flags().GetString(linkFlagName)
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
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

	lkc, err := getKafkaClusterLkcId(c.AuthenticatedCLICommand)
	if err != nil {
		return err
	}

	promoteMirrorOpt := &kafkarestv3.UpdateKafkaMirrorTopicsPromoteOpts{
		AlterMirrorsRequestData: optional.NewInterface(kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: args}),
		ValidateOnly:            optional.NewBool(dryRun),
	}

	results, httpResp, err := kafkaREST.Client.ClusterLinkingV3Api.UpdateKafkaMirrorTopicsPromote(kafkaREST.Context, lkc, linkName, promoteMirrorOpt)
	if err != nil {
		return kafkarest.NewError(kafkaREST.CloudClient.GetUrl(), err, httpResp)
	}

	return printAlterMirrorResult(cmd, results)
}
