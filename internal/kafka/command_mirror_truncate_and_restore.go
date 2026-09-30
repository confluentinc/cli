package kafka

import (
	"github.com/spf13/cobra"

	kafkarestv3 "github.com/confluentinc/ccloud-sdk-go-v2/kafkarest/v3"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *mirrorCommand) newTruncateAndRestoreCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "truncate-and-restore <local-topic-1> [local-topic-2] ... [local-topic-N]",
		Short:             "Truncate topics and restore mirroring.",
		RunE:              c.truncateAndRestore,
		Args:              cobra.MinimumNArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgsMultiple),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Truncates and restores local topics for "my-topic-1" and "my-topic-2":`,
				Code: "confluent kafka mirror truncate-and-restore my-topic-1 my-topic-2 --link my-link",
			},
		),
	}

	pcmd.AddLinkFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().Bool(dryrunFlagName, false, "If set, does not actually truncate the local topic, but simply validates it.")
	cmd.Flags().Bool(includePartitionDataFlagName, false, "If set, returns the number of messages truncated per partition.")
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("link"))

	return cmd
}

func (c *mirrorCommand) truncateAndRestore(cmd *cobra.Command, args []string) error {
	link, err := cmd.Flags().GetString("link")
	if err != nil {
		return err
	}

	dryRun, err := cmd.Flags().GetBool(dryrunFlagName)
	if err != nil {
		return err
	}

	includePartitionData, err := cmd.Flags().GetBool(includePartitionDataFlagName)
	if err != nil {
		return err
	}

	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	data := kafkarestv3.AlterMirrorsRequestData{MirrorTopicNames: &args}

	results, err := kafkaREST.CloudClient.UpdateKafkaTruncateAndRestore(link, dryRun, includePartitionData, data)
	if err != nil {
		return err
	}

	return printAlterMirrorResult(cmd, results)
}
