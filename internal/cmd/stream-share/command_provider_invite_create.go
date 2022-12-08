package streamshare

import (
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

	deliveryMethod := "Email"
	topicCRN, err := getTopicCRN(c.Config.GetLastUsedOrgId(), environment, srCluster.Id, kafkaCluster, topic)
	if err != nil {
		return err
	}

	subjectsCrn := make([]string, 0, len(schemaRegistrySubjects))
	for _, subject := range schemaRegistrySubjects {
		crn, err := getSubjectCRN(c.Config.GetLastUsedOrgId(), environment, srCluster.Id, subject)
		if err != nil {
			return err
		}
		subjectsCrn = append(subjectsCrn, crn)
	}

	err = c.validateSubjects(subjectsCrn, topicCRN)
	if err != nil {
		return err
	}

	resources := []string{topicCRN}
	resources = append(resources, subjectsCrn...)

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

func (c *command) validateSubjects(newSubjectsCRN []string, topicCRN string) error {
	providerShares, err := c.V2Client.ListProviderShares("", topicCRN)
	if err != nil {
		return err
	}

	if len(providerShares) == 0 {
		return nil
	}

	sharedResources, err := c.V2Client.ListProviderSharedResources(topicCRN)
	if err != nil {
		return err
	}

	existingSubjectsCRN, err := getSubjectsCRNFromSharedResources(sharedResources)
	if err != nil {
		return err
	}

	return areSubjectsModified(newSubjectsCRN, existingSubjectsCRN)
}
