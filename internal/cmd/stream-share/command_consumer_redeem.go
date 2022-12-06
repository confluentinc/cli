package streamshare

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	redeemTokenFields            = []string{"Id", "ApiKey", "ApiSecret", "KafkaBootstrapUrl", "SchemaRegistryApiKey", "SchemaRegistrySecret", "SchemaRegistryUrl", "Resources"}
	redeemTokenPrivateLinkFields = []string{"NetworkDnsDomain", "NetworkZones", "NetworkZonalSubdomains", "NetworkKind",
		"NetworkPrivateLinkDataType", "NetworkPrivateLinkData"}
	redeemTokenHumanLabelMap = map[string]string{
		"Id":                   "ID",
		"ApiKey":               "API Key",
		"ApiSecret":            "API Secret",
		"KafkaBootstrapUrl":    "Kafka Bootstrap URL",
		"SchemaRegistryApiKey": "Schema Registry API Key",
		"SchemaRegistrySecret": "Schema Registry Secret",
		"SchemaRegistryUrl":    "Schema Registry URL",
		"Resources":            "Resources",
	}
	redeemTokenPrivateLinkHumanLabelMap = map[string]string{
		"NetworkDnsDomain":           "Network DNS Domain",
		"NetworkZones":               "Network Zones",
		"NetworkZonalSubdomains":     "Network Zonal Subdomains",
		"NetworkKind":                "Network Kind",
		"NetworkPrivateLinkDataType": "Network Private Link Data Type",
		"NetworkPrivateLinkData":     "Network Private Link Data",
	}
	redeemTokenStructuredLabelMap = map[string]string{
		"Id":                   "id",
		"ApiKey":               "api_key",
		"ApiSecret":            "secret",
		"KafkaBootstrapUrl":    "kafka_bootstrap_url",
		"SchemaRegistryApiKey": "schema_registry_api_key",
		"SchemaRegistrySecret": "schema_registry_secret",
		"SchemaRegistryUrl":    "schema_registry_url",
		"Resources":            "resources",
	}
	redeemTokenPrivateLinkStructuredLabelMap = map[string]string{
		"NetworkDnsDomain":           "network_dns_domain",
		"NetworkZones":               "network_zones",
		"NetworkZonalSubdomains":     "network_zonal_subdomains",
		"NetworkKind":                "network_kind",
		"NetworkPrivateLinkDataType": "network_private_link_data_type",
		"NetworkPrivateLinkData":     "network_private_link_data",
	}
)

type privateLinkNetworkDetails struct {
	networkKind         string
	privateLinkDataType string
	privateLinkData     interface{}
}

type redeemToken struct {
	Id                         string
	ApiKey                     string
	ApiSecret                  string
	KafkaBootstrapUrl          string
	SchemaRegistryApiKey       string
	SchemaRegistrySecret       string
	SchemaRegistryUrl          string
	Resources                  []string
	NetworkDnsDomain           string
	NetworkZones               string
	NetworkZonalSubdomains     []string
	NetworkKind                string
	NetworkPrivateLinkDataType string
	NetworkPrivateLinkData     interface{}
}

func (c *command) newRedeemCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redeem <token>",
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

	cmd.Flags().String("aws-account-id", "", "Consumer's AWS account ID for PrivateLink access.")
	cmd.Flags().String("azure-subscription-id", "", "Consumer's Azure subscription ID for PrivateLink access.")
	cmd.Flags().String("gcp-project-id", "", "Consumer's GCP project ID for Private Service Connect access.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) redeemShare(cmd *cobra.Command, args []string) error {
	token := args[0]

	awsAccountId, err := cmd.Flags().GetString("aws-account-id")
	if err != nil {
		return err
	}

	azureSubscriptionId, err := cmd.Flags().GetString("azure-subscription-id")
	if err != nil {
		return err
	}

	gcpProjectId, err := cmd.Flags().GetString("gcp-project-id")
	if err != nil {
		return err
	}

	redeemResponse, err := c.V2Client.RedeemSharedToken(token, awsAccountId, azureSubscriptionId, gcpProjectId)
	if err != nil {
		return err
	}

	var resources []string
	for _, resource := range redeemResponse.GetResources() {
		if resource.CdxV1SharedTopic != nil {
			resources = append(resources, fmt.Sprintf(`%s="%s"`, resource.CdxV1SharedTopic.GetKind(), resource.CdxV1SharedTopic.GetTopic()))
		}

		if resource.CdxV1SharedGroup != nil {
			resources = append(resources, fmt.Sprintf(`Group_Prefix="%s"`, resource.CdxV1SharedGroup.GetGroupPrefix()))
		}
	}

	tokenObj := &redeemToken{
		Id:                   redeemResponse.GetId(),
		ApiKey:               redeemResponse.GetApiKey(),
		ApiSecret:            redeemResponse.GetSecret(),
		KafkaBootstrapUrl:    redeemResponse.GetKafkaBootstrapUrl(),
		SchemaRegistryApiKey: redeemResponse.GetSchemaRegistryApiKey(),
		SchemaRegistrySecret: redeemResponse.GetSchemaRegistrySecret(),
		SchemaRegistryUrl:    redeemResponse.GetSchemaRegistryUrl(),
		Resources:            resources,
	}

	// non private link cluster
	if awsAccountId == "" && azureSubscriptionId == "" && gcpProjectId == "" {
		return output.DescribeObject(cmd, tokenObj, redeemTokenFields, redeemTokenHumanLabelMap, redeemTokenStructuredLabelMap)
	}

	return c.handlePrivateLinkClusterRedeem(cmd, redeemResponse, tokenObj)
}

func (c *command) handlePrivateLinkClusterRedeem(cmd *cobra.Command, redeemResponse cdxv1.CdxV1RedeemTokenResponse, tokenObj *redeemToken) error {
	consumerSharedResources, err := c.V2Client.ListConsumerSharedResources(redeemResponse.GetId())
	if err != nil {
		return err
	}

	var network cdxv1.CdxV1Network
	if len(consumerSharedResources) != 0 {
		privateNetwork, err := c.V2Client.GetPrivateLinkNetworkConfig(consumerSharedResources[0].GetId())
		if err != nil {
			return err
		}
		network = privateNetwork
	}

	networkDetails := getPrivateLinkNetworkDetails(network)
	tokenObj.NetworkDnsDomain = network.GetDnsDomain()
	tokenObj.NetworkZones = strings.Join(network.GetZones(), ",")
	tokenObj.NetworkZonalSubdomains = mapSubdomainsToList(network.GetZonalSubdomains())
	tokenObj.NetworkKind = networkDetails.networkKind
	tokenObj.NetworkPrivateLinkDataType = networkDetails.privateLinkDataType
	tokenObj.NetworkPrivateLinkData = networkDetails.privateLinkData

	return output.DescribeObject(cmd, tokenObj, append(redeemTokenFields, redeemTokenPrivateLinkFields...),
		combineMaps(redeemTokenHumanLabelMap, redeemTokenPrivateLinkHumanLabelMap),
		combineMaps(redeemTokenStructuredLabelMap, redeemTokenPrivateLinkStructuredLabelMap))
}
