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
				Code: "confluent stream-share provider invite create --email user@example.com --topic topic-12345 --environment env-12345 --cluster lkc-12345",
			},
		),
	}

	cmd.Flags().String("email", "", "Email of the user with whom to share the topic.")
	cmd.Flags().String("topic", "", "Topic to be shared.")
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("schema-registry-subjects", []string{}, "A comma-separated list of Schema Registry subjects.")
	pcmd.AddOutputFlag(cmd)

	_ = cmd.MarkFlagRequired("email")
	_ = cmd.MarkFlagRequired("topic")
	_ = cmd.MarkFlagRequired("environment")
	_ = cmd.MarkFlagRequired("cluster")

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

	schemaRegistrySubjects, err := cmd.Flags().GetStringSlice("schema-registry-subjects")
	if err != nil {
		return err
	}

	srCluster, err := c.Context.FetchSchemaRegistryByAccountId(cmd.Context(), environment)
	if err != nil {
		return err
	}

	invite, err := c.V2Client.CreateProviderInvite(environment, kafkaCluster, topic, email, srCluster.Id,
		c.Config.GetLastUsedOrgId(), schemaRegistrySubjects)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(c.buildProviderShare(invite))
	return table.Print()
}
