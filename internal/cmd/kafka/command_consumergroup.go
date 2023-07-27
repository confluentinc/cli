package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

type consumerGroupCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type consumerGroupOut struct {
	ClusterId         string `human:"Cluster" serialized:"cluster"`
	ConsumerGroupId   string `human:"Consumer Group" serialized:"consumer_group"`
	Coordinator       string `human:"Coordinator" serialized:"coordinator"`
	IsSimple          bool   `human:"Simple" serialized:"simple"`
	PartitionAssignor string `human:"Partition Assignor" serialized:"partition_assignor"`
	State             string `human:"State" serialized:"state"`
}

func newConsumerGroupCommand(cfg *v1.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "consumer-group",
		Aliases:     []string{"cg"},
		Short:       "Manage Kafka consumer groups.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLoginOrOnPremLogin},
	}

	c := &consumerGroupCommand{}

	if cfg.IsCloudLogin() {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedCLICommand(cmd, prerunner)

		cmd.AddCommand(c.newDescribeCommand())
		cmd.AddCommand(newLagCommand(prerunner))
		cmd.AddCommand(c.newListCommand())
	} else {
		c.AuthenticatedCLICommand = pcmd.NewAuthenticatedWithMDSCLICommand(cmd, prerunner)
		c.PersistentPreRunE = prerunner.InitializeOnPremKafkaRest(c.AuthenticatedCLICommand)

		cmd.AddCommand(c.newDescribeCommandOnPrem())
		cmd.AddCommand(c.newListCommandOnPrem())
	}

	return cmd
}

func (c *consumerGroupCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteConsumerGroups(c.AuthenticatedCLICommand)
}
