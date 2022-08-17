package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newCreateEmailInviteCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Invite a consumer with email.",
		Args:  cobra.NoArgs,
		RunE:  c.createEmailInvite,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Invite a user with email "user@example.com":`,
				Code: "confluent stream-share provider invite create --email user@example.com --environment env-12345 --kafka-cluster lkc-12345 --topic topic-12345",
			},
		),
	}

	cmd.Flags().String("email", "", "Email of the user with whom the topic is shared")
	cmd.Flags().String("topic", "", "Topic to be shared")
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("environment")
	_ = cmd.MarkFlagRequired("cluster")
	_ = cmd.MarkFlagRequired("topic")

	return cmd
}

func (c *command) createEmailInvite(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	kafkaCluster, err := cmd.Flags().GetString("cluster")
	if err != nil {
		return err
	}

	topic, err := cmd.Flags().GetString("topic")
	if err != nil {
		return err
	}

	email, err := cmd.Flags().GetString("email")
	if err != nil {
		return err
	}

	invite, _, err := c.V2Client.CreateInvite(environment, kafkaCluster, topic, email)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, c.buildProviderShare(invite), providerShareListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
