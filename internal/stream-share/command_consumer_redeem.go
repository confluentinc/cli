package streamshare

import (
	"fmt"

	"github.com/spf13/cobra"

	cdxv1 "github.com/confluentinc/ccloud-sdk-go-v2/cdx/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type redeemOut struct {
	Id                         string            `human:"ID" serialized:"id"`
	ApiKey                     string            `human:"API Key" serialized:"api_key"`
	ApiSecret                  string            `human:"API Secret" serialized:"api_secret"`
	KafkaBootstrapUrl          string            `human:"Kafka Bootstrap URL" serialized:"kafka_bootstrap_url"`
	SchemaRegistryApiKey       string            `human:"Schema Registry API Key" serialized:"schema_registry_api_key"`
	SchemaRegistryAPISecret    string            `human:"Schema Registry API Secret" serialized:"schema_registry_api_secret"`
	SchemaRegistryUrl          string            `human:"Schema Registry URL" serialized:"schema_registry_url"`
	Resources                  []string          `human:"Resources" serialized:"resources"`
	NetworkDnsDomain           string            `human:"Network DNS Domain" serialized:"network_dns_domain"`
	NetworkZones               []string          `human:"Network Zones" serialized:"network_zones"`
	NetworkZonalSubdomains     map[string]string `human:"Network Zonal Subdomains" serialized:"network_zonal_subdomains"`
	NetworkKind                string            `human:"Network Kind" serialized:"network_kind"`
	NetworkPrivateLinkDataType string            `human:"Network Private Link Data Type" serialized:"network_private_link_data_type"`
	NetworkPrivateLinkData     string            `human:"Network Private Link Data" serialized:"network_private_link_data"`
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

	cmd.Flags().String("aws-account", "", "Consumer's AWS account ID for PrivateLink access.")
	cmd.Flags().String("azure-subscription", "", "Consumer's Azure subscription ID for PrivateLink access.")
	cmd.Flags().String("gcp-project", "", "Consumer's GCP project ID for Private Service Connect access.")
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) redeemShare(cmd *cobra.Command, args []string) error {
	awsAccount, err := cmd.Flags().GetString("aws-account")
	if err != nil {
		return err
	}

	azureSubscription, err := cmd.Flags().GetString("azure-subscription")
	if err != nil {
		return err
	}

	gcpProject, err := cmd.Flags().GetString("gcp-project")
	if err != nil {
		return err
	}

	redeemResponse, err := c.V2Client.RedeemSharedToken(args[0], awsAccount, azureSubscription, gcpProject)
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

	isPrivateLink := awsAccount != "" || azureSubscription != "" || gcpProject != ""

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

		table.Add(&redeemOut{
			Id:                         redeemResponse.GetId(),
			ApiKey:                     redeemResponse.GetApiKey(),
			ApiSecret:                  redeemResponse.GetSecret(),
			KafkaBootstrapUrl:          redeemResponse.GetKafkaBootstrapUrl(),
			SchemaRegistryApiKey:       redeemResponse.GetSchemaRegistryApiKey(),
			SchemaRegistryAPISecret:    redeemResponse.GetSchemaRegistrySecret(),
			SchemaRegistryUrl:          redeemResponse.GetSchemaRegistryUrl(),
			Resources:                  resources,
			NetworkDnsDomain:           network.GetDnsDomain(),
			NetworkZones:               network.GetZones(),
			NetworkZonalSubdomains:     network.GetZonalSubdomains(),
			NetworkKind:                networkDetails.networkKind,
			NetworkPrivateLinkDataType: networkDetails.privateLinkDataType,
			NetworkPrivateLinkData:     networkDetails.privateLinkData,
		})
	} else {
		table.Add(&redeemOut{
			Id:                      redeemResponse.GetId(),
			ApiKey:                  redeemResponse.GetApiKey(),
			ApiSecret:               redeemResponse.GetSecret(),
			KafkaBootstrapUrl:       redeemResponse.GetKafkaBootstrapUrl(),
			SchemaRegistryApiKey:    redeemResponse.GetSchemaRegistryApiKey(),
			SchemaRegistryAPISecret: redeemResponse.GetSchemaRegistrySecret(),
			SchemaRegistryUrl:       redeemResponse.GetSchemaRegistryUrl(),
			Resources:               resources,
		})

		table.Filter([]string{"Id", "ApiKey", "ApiSecret", "KafkaBootstrapUrl", "SchemaRegistryApiKey", "SchemaRegistryAPISecret", "SchemaRegistryUrl", "Resources"})
	}

	return table.Print()
}
