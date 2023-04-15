package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

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

	cobra.CheckErr(cmd.MarkFlagRequired("email"))
	cobra.CheckErr(cmd.MarkFlagRequired("topic"))
	cobra.CheckErr(cmd.MarkFlagRequired("environment"))
	cobra.CheckErr(cmd.MarkFlagRequired("cluster"))

	return cmd
}

func (c *command) createEmailInvite(cmd *cobra.Command, _ []string) error {
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}

	cluster, err := cmd.Flags().GetString("cluster")
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

	srCluster, err := c.Context.FetchSchemaRegistryByEnvironmentId(cmd.Context(), environment)
	if err != nil {
		return err
	}

	deliveryMethod := "Email"
	resources := []string{
		fmt.Sprintf("crn://confluent.cloud/organization=%s/environment=%s/schema-registry=%s/kafka=%s/topic=%s",
			c.Context.GetCurrentOrganization(), environment, srCluster.Id, cluster, topic),
	}
	for _, subject := range schemaRegistrySubjects {
		resources = append(resources, fmt.Sprintf("crn://confluent.cloud/organization=%s/environment=%s/schema-registry=%s/subject=%s",
			c.Context.GetCurrentOrganization(), environment, srCluster.Id, subject))
	}

	shareReq := cdxv1.CdxV1CreateProviderShareRequest{
		ConsumerRestriction: &cdxv1.CdxV1CreateProviderShareRequestConsumerRestrictionOneOf{
			CdxV1EmailConsumerRestriction: &cdxv1.CdxV1EmailConsumerRestriction{
				Kind:  deliveryMethod,
				Email: email,
			},
		},
		DeliveryMethod: &deliveryMethod,
		Resources:      &resources,
	}

	invite, err := c.V2Client.CreateProviderInvite(shareReq)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(c.buildProviderShare(invite))
	return table.Print()
}
