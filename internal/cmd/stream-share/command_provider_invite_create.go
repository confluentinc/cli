package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (s *inviteCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a consumer invite based on email.",
		Args:  cobra.NoArgs,
		RunE:  s.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create an email invite for user with email user@example.com:",
				Code: "confluent stream-share provider invite create --email user@example.com --environment env-12345 --kafka-cluster lkc-12345 --topic topic-12345",
			},
		),
	}

	cmd.Flags().String("email", "", "Email of the user with whom the topic is shared")
	cmd.Flags().String("environment", "", "ID of ccloud environment")
	cmd.Flags().String("kafka-cluster", "", "ID of the Kafka cluster")
	cmd.Flags().String("topic", "", "Topic to be shared")

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("environment")
	_ = cmd.MarkFlagRequired("kafka-cluster")
	_ = cmd.MarkFlagRequired("topic")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (s *inviteCommand) create(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	kafkaCluster, err := cmd.Flags().GetString("kafka-cluster")
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

	invite, _, err := s.V2Client.CreateInvite(environment, kafkaCluster, topic, email)
	if err != nil {
		return err
	}

	return output.DescribeObject(cmd, buildProviderShare(invite), providerShareListFields, providerHumanLabelMap, providerStructuredLabelMap)
}
