package streamshare

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type redeemHumanOut struct {
	Id                         string `human:"ID"`
	ApiKey                     string `human:"API Key"`
	ApiSecret                  string `human:"API Secret"`
	KafkaBootstrapUrl          string `human:"Kafka Bootstrap URL"`
	SchemaRegistryApiKey       string `human:"Schema Registry API Key"`
	SchemaRegistrySecret       string `human:"Schema Registry Secret"`
	SchemaRegistryUrl          string `human:"Schema Registry URL"`
	Resources                  string `human:"Resources"`
	NetworkDnsDomain           string `human:"Network DNS Domain"`
	NetworkZones               string `human:"Network Zones"`
	NetworkZonalSubdomains     string `human:"Network Zonal Subdomains"`
	NetworkKind                string `human:"Network Kind"`
	NetworkPrivateLinkDataType string `human:"Network Private Link Data Type"`
	NetworkPrivateLinkData     string `human:"Network Private Link Data"`
}

type redeemSerializedOut struct {
	Id                         string   `serialized:"id"`
	ApiKey                     string   `serialized:"api_key"`
	ApiSecret                  string   `serialized:"api_secret"`
	KafkaBootstrapUrl          string   `serialized:"kafka_bootstrap_url"`
	SchemaRegistryApiKey       string   `serialized:"schema_registry_api_key"`
	SchemaRegistrySecret       string   `serialized:"schema_registry_secret"`
	SchemaRegistryUrl          string   `serialized:"schema_registry_url"`
	Resources                  []string `serialized:"resources"`
	NetworkDnsDomain           string   `serialized:"network_dns_domain"`
	NetworkZones               string   `serialized:"network_zones"`
	NetworkZonalSubdomains     []string `serialized:"network_zonal_subdomains"`
	NetworkKind                string   `serialized:"network_kind"`
	NetworkPrivateLinkDataType string   `serialized:"network_private_link_data_type"`
	NetworkPrivateLinkData     string   `serialized:"network_private_link_data"`
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

	redeemResponse, err := c.V2Client.RedeemSharedToken(args[0], awsAccountId, azureSubscriptionId, gcpProjectId)
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

	isPrivateLink := awsAccountId != "" || azureSubscriptionId != "" || gcpProjectId != ""

	table := output.NewTable(cmd)
	if isPrivateLink {
		consumerSharedResources, err := c.V2Client.ListConsumerSharedResources(redeemResponse.GetId())
		if err != nil {
			return err
		}

		var network cdxv1.CdxV1Network
		if len(consumerSharedResources) > 0 {
			privateNetwork, err := c.V2Client.GetPrivateLinkNetworkConfig(consumerSharedResources[0].GetId())
			if err != nil {
				return err
			}
			network = privateNetwork
		}
		networkDetails := getPrivateLinkNetworkDetails(network)

		if output.GetFormat(cmd) == output.Human {
			table.Add(&redeemHumanOut{
				Id:                         redeemResponse.GetId(),
				ApiKey:                     redeemResponse.GetApiKey(),
				ApiSecret:                  redeemResponse.GetSecret(),
				KafkaBootstrapUrl:          redeemResponse.GetKafkaBootstrapUrl(),
				SchemaRegistryApiKey:       redeemResponse.GetSchemaRegistryApiKey(),
				SchemaRegistrySecret:       redeemResponse.GetSchemaRegistrySecret(),
				SchemaRegistryUrl:          redeemResponse.GetSchemaRegistryUrl(),
				Resources:                  strings.Join(resources, ", "),
				NetworkDnsDomain:           network.GetDnsDomain(),
				NetworkZones:               strings.Join(network.GetZones(), ", "),
				NetworkZonalSubdomains:     strings.Join(mapSubdomainsToList(network.GetZonalSubdomains()), ", "),
				NetworkKind:                networkDetails.networkKind,
				NetworkPrivateLinkDataType: networkDetails.privateLinkDataType,
				NetworkPrivateLinkData:     networkDetails.privateLinkData,
			})
		} else {
			table.Add(&redeemSerializedOut{
				Id:                   redeemResponse.GetId(),
				ApiKey:               redeemResponse.GetApiKey(),
				ApiSecret:            redeemResponse.GetSecret(),
				KafkaBootstrapUrl:    redeemResponse.GetKafkaBootstrapUrl(),
				SchemaRegistryApiKey: redeemResponse.GetSchemaRegistryApiKey(),
				SchemaRegistrySecret: redeemResponse.GetSchemaRegistrySecret(),
				SchemaRegistryUrl:    redeemResponse.GetSchemaRegistryUrl(),
				Resources:            resources,
				NetworkDnsDomain:     network.GetDnsDomain(),
				// TODO: Serialize array instead of string in next major version
				NetworkZones:               strings.Join(network.GetZones(), ","),
				NetworkZonalSubdomains:     mapSubdomainsToList(network.GetZonalSubdomains()),
				NetworkKind:                networkDetails.networkKind,
				NetworkPrivateLinkDataType: networkDetails.privateLinkDataType,
				NetworkPrivateLinkData:     networkDetails.privateLinkData,
			})
		}
	} else {
		if output.GetFormat(cmd) == output.Human {
			table.Add(&redeemHumanOut{
				Id:                   redeemResponse.GetId(),
				ApiKey:               redeemResponse.GetApiKey(),
				ApiSecret:            redeemResponse.GetSecret(),
				KafkaBootstrapUrl:    redeemResponse.GetKafkaBootstrapUrl(),
				SchemaRegistryApiKey: redeemResponse.GetSchemaRegistryApiKey(),
				SchemaRegistrySecret: redeemResponse.GetSchemaRegistrySecret(),
				SchemaRegistryUrl:    redeemResponse.GetSchemaRegistryUrl(),
				Resources:            strings.Join(resources, ", "),
			})
		} else {
			table.Add(&redeemSerializedOut{
				Id:                   redeemResponse.GetId(),
				ApiKey:               redeemResponse.GetApiKey(),
				ApiSecret:            redeemResponse.GetSecret(),
				KafkaBootstrapUrl:    redeemResponse.GetKafkaBootstrapUrl(),
				SchemaRegistryApiKey: redeemResponse.GetSchemaRegistryApiKey(),
				SchemaRegistrySecret: redeemResponse.GetSchemaRegistrySecret(),
				SchemaRegistryUrl:    redeemResponse.GetSchemaRegistryUrl(),
				Resources:            resources,
			})
		}

		table.Filter([]string{"Id", "ApiKey", "ApiSecret", "KafkaBootstrapUrl", "SchemaRegistryApiKey", "SchemaRegistrySecret", "SchemaRegistryUrl", "Resources"})
	}

	return table.Print()
}
