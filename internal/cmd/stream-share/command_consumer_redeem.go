package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	redeemTokenFields        = []string{"Id", "ApiKey", "Secret", "KafkaBootstrapUrl", "Resources"}
	redeemTokenHumanLabelMap = map[string]string{
		"Id":                "ID",
		"ApiKey":            "API Key",
		"Secret":            "Secret",
		"KafkaBootstrapUrl": "Kafka Bootstrap URL",
		"Resources":         "Resources",
	}
	redeemTokenStructuredLabelMap = map[string]string{
		"Id":                "id",
		"ApiKey":            "api_key",
		"Secret":            "secret",
		"KafkaBootstrapUrl": "kafka_bootstrap_url",
		"Resources":         "resources",
	}
)

type redeemToken struct {
	Id                string
	ApiKey            string
	Secret            string
	KafkaBootstrapUrl string
	Resources         []string
}

func (c *command) newRedeemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem <stream-share-token>",
		Short: "Redeem a stream share token.",
		RunE:  c.redeemShare,
		Args:  cobra.ExactArgs(1),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Redeem a stream share token:`,
				Code: "confluent stream-share consumer redeem DBBG8xGRfh85ePuk4x5BaENvb25vaGsydXdhejRVNp-pOzCWOLF85LzqcZCq1lVe8OQxSJqQo8XgUMRbtVs5fqbpM5BUKhnHAUcd3C5ip_yWfd3BFRlMVxGQwYo75aSQDb44ACdoAcgjwLH_9YVbk4GJoK-BtZtlpjYSTAIBbhvbFWWOU1bcFyW3HetlyzTIlIjG_UkSKFfDZ_5YNNuw0CBLZQf14J36b4QpSLe05jx9s695tINCm-dyPLX8_pUIqA2ekEZyf86pE7Azh7NBZz00uGZ0FrRl_ir9UvHF1uZ9sID6aZc=",
			},
		),
	}

	cmd.Flags().String("aws-account-id", "", "The AWS account ID for the consumer network.")
	cmd.Flags().String("azure-subscription-id", "", "The Azure subscription ID for the consumer network.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) redeemShare(cmd *cobra.Command, args []string) error {
	token := args[0]

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

	tokenObj := &redeemToken{
		Id:                redeemResponse.GetId(),
		ApiKey:            redeemResponse.GetApikey(),
		Secret:            redeemResponse.GetSecret(),
		KafkaBootstrapUrl: redeemResponse.GetKafkaBootstrapUrl(),
		Resources:         resources,
	}

	return output.DescribeObject(cmd, tokenObj, redeemTokenFields, redeemTokenHumanLabelMap, redeemTokenStructuredLabelMap)
}
