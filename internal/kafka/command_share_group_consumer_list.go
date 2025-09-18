package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type shareGroupConsumerOut struct {
	Cluster    string `human:"Cluster" serialized:"cluster"`
	ShareGroup string `human:"Share Group" serialized:"share_group"`
	Consumer   string `human:"Consumer" serialized:"consumer"`
	Client     string `human:"Client" serialized:"client"`
}

func (c *shareCommand) newGroupConsumerListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Kafka share group consumers.",
		Args:  cobra.NoArgs,
		RunE:  c.groupConsumerList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List all consumers in share group "my-share-group".`,
				Code: "confluent kafka share group consumer list --group my-share-group",
			},
		),
	}

	c.addShareGroupFlag(cmd)
	pcmd.AddEndpointFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("group"))

	return cmd
}

func (c *shareCommand) groupConsumerList(cmd *cobra.Command, _ []string) error {
	kafkaREST, err := c.GetKafkaREST(cmd)
	if err != nil {
		return err
	}

	group, err := cmd.Flags().GetString("group")
	if err != nil {
		return err
	}

	consumers, err := kafkaREST.CloudClient.ListKafkaShareGroupConsumers(group)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, consumer := range consumers {
		list.Add(&shareGroupConsumerOut{
			Cluster:    consumer.GetClusterId(),
			ShareGroup: group, // Use the group ID from the command flag
			Consumer:   consumer.GetConsumerId(),
			Client:     consumer.GetClientId(),
		})
	}
	return list.Print()
}

func (c *shareCommand) addShareGroupFlag(cmd *cobra.Command) {
	cmd.Flags().String("group", "", "Share group ID.")

	pcmd.RegisterFlagCompletionFunc(cmd, "group", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		return pcmd.AutocompleteShareGroups(cmd, c.AuthenticatedCLICommand)
	})
}
