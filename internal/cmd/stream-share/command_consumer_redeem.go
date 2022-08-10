package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	redeemTokenFields        = []string{"Id", "APIKey", "Secret", "KafkaBootstrapURL", "Resources"}
	redeemTokenHumanLabelMap = map[string]string{
		"Id":                "ID",
		"APIKey":            "API Key",
		"Secret":            "Secret",
		"KafkaBootstrapURL": "Kafka Bootstrap URL",
		"Resources":         "Resources",
	}
	redeemTokenStructuredLabelMap = map[string]string{
		"Id":                "id",
		"APIKey":            "apikey",
		"Secret":            "secret",
		"KafkaBootstrapURL": "kafka_bootstrap_url",
		"Resources":         "resources",
	}

	redeemPreviewFields = []string{"Id", "Cloud", "DisplayName", "Description", "OrganizationName", "OrganizationDetails",
		"OrganizationContact", "NetworkConnectionTypes", "Labels"}
	redeemPreviewHumanLabels = []string{"ID", "Cloud", "Display Name", "Description", "Organization Name", "Organization Details",
		"Organization Contact", "Network Connection Types", "Labels"}
	redeemPreviewStructuredLabels = []string{"id", "cloud", "display_name", "description", "organization_name",
		"organization_details", "organization_contact", "network_connection_types", "labels"}
)

type redeemToken struct {
	Id                string
	APIKey            string
	Secret            string
	KafkaBootstrapURL string
	Resources         []string
}

type redeemPreview struct {
	Id                     string
	Cloud                  string
	DisplayName            string
	Description            string
	OrganizationName       string
	OrganizationDetails    string
	OrganizationContact    string
	NetworkConnectionTypes []string
	Labels                 []string
}

func (c *consumerCommand) newRedeemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem <stream-share-token>",
		Short: "Redeem stream share token.",
		RunE:  c.redeem,
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Redeem stream share token "stream-share-token":`,
				Code: "confluent stream-share consumer redeem stream-share-token",
			},
		),
	}

	cmd.Flags().Bool("preview", false, "Preview shared resources without redeeming the token.")
	cmd.Flags().String("aws-account", "", "The AWS account id for the consumer network.")
	cmd.Flags().String("azure-subscription", "", "The Azure subscription for the consumer network.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *consumerCommand) redeem(cmd *cobra.Command, args []string) error {
	isPreview, err := cmd.Flags().GetBool("preview")
	if err != nil {
		return err
	}

	token := args[0]

	if isPreview {
		return c.handleRedeemPreview(cmd, token)
	}

	return c.handleRedeem(cmd, token)
}

func (c *consumerCommand) handleRedeem(cmd *cobra.Command, token string) error {
	awsAccount, err := cmd.Flags().GetString("aws-account")
	if err != nil {
		return err
	}

	azureSubscription, err := cmd.Flags().GetString("azure-subscription")
	if err != nil {
		return err
	}

	redeemResponse, _, err := c.V2Client.RedeemSharedToken(token, awsAccount, azureSubscription)
	if err != nil {
		return err
	}

	var resources []string

	for _, resource := range redeemResponse.GetResources() {
		resources = append(resources, fmt.Sprintf("%s:%s", resource.CdxV1SharedTopic.Kind, resource.CdxV1SharedTopic.Topic))
	}

	return output.DescribeObject(cmd, &redeemToken{
		Id:                redeemResponse.GetId(),
		APIKey:            redeemResponse.GetApikey(),
		Secret:            redeemResponse.GetSecret(),
		KafkaBootstrapURL: redeemResponse.GetKafkaBootstrapUrl(),
		Resources:         resources,
	}, redeemTokenFields, redeemTokenHumanLabelMap, redeemTokenStructuredLabelMap)
}

func (c *consumerCommand) handleRedeemPreview(cmd *cobra.Command, token string) error {
	tokenPreview, _, err := c.V2Client.PreviewSharedToken(token)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, redeemPreviewFields, redeemPreviewHumanLabels, redeemPreviewStructuredLabels)
	if err != nil {
		return err
	}

	for _, resource := range *tokenPreview.ConsumerResources {
		element := &redeemPreview{
			Id:                     resource.GetId(),
			Cloud:                  resource.GetCloud(),
			DisplayName:            resource.GetDisplayName(),
			Description:            resource.GetDescription(),
			OrganizationName:       resource.GetOrganizationName(),
			OrganizationDetails:    resource.GetOrganizationDetails(),
			OrganizationContact:    resource.GetOrganizationContact(),
			NetworkConnectionTypes: resource.GetNetworkConnectionTypes().Items,
			Labels:                 resource.GetLabels(),
		}

		outputWriter.AddElement(element)
	}

	return outputWriter.Out()
}
