package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkinggatewayv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-gateway/v1"

	"github.com/confluentinc/cli/v4/pkg/ccloudv2"
	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/network"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const (
	awsEgressPrivateLink           = "AwsEgressPrivateLink"
	awsIngressPrivateLink          = "AwsIngressPrivateLink"
	awsPeering                     = "AwsPeering"
	azureEgressPrivateLink         = "AzureEgressPrivateLink"
	azurePeering                   = "AzurePeering"
	awsPrivateNetworkInterface     = "AwsPrivateNetworkInterface"
	gcpPeering                     = "GcpPeering"
	gcpEgressPrivateServiceConnect = "GcpEgressPrivateServiceConnect"
)

var (
	createGatewayTypes = []string{"egress-privatelink", "ingress-privatelink", "private-network-interface"}
	listGatewayTypes   = []string{"aws-egress-privatelink", "aws-ingress-privatelink", "azure-egress-privatelink", "gcp-egress-private-service-connect"} // TODO: check if we accept private-network-interface here
	gatewayTypeMap     = map[string]string{
		"aws-egress-privatelink":             awsEgressPrivateLink,
		"aws-ingress-privatelink":            awsIngressPrivateLink,
		"azure-egress-privatelink":           azureEgressPrivateLink,
		"gcp-egress-private-service-connect": gcpEgressPrivateServiceConnect,
	}
)

type gatewayOut struct {
	Id                string   `human:"ID" serialized:"id"`
	Name              string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment       string   `human:"Environment" serialized:"environment"`
	Region            string   `human:"Region,omitempty" serialized:"region,omitempty"`
	Type              string   `human:"Type,omitempty" serialized:"type,omitempty"`
	AwsPrincipalArn   string   `human:"AWS Principal ARN,omitempty" serialized:"aws_principal_arn,omitempty"`
	AzureSubscription string   `human:"Azure Subscription,omitempty" serialized:"azure_subscription,omitempty"`
	GcpIamPrincipal   string   `human:"GCP IAM Principal,omitempty" serialized:"gcp_iam_principal,omitempty"`
	GcpProject        string   `human:"GCP Project,omitempty" serialized:"gcp_project,omitempty"`
	Phase             string   `human:"Phase" serialized:"phase"`
	Zones             []string `human:"Zones,omitempty" serialized:"zones,omitempty"`
	Account           string   `human:"Account,omitempty" serialized:"account,omitempty"`
	ErrorMessage      string   `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
}

func (c *command) newGatewayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage network gateways.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newGatewayCreateCommand())
	cmd.AddCommand(c.newGatewayDeleteCommand())
	cmd.AddCommand(c.newGatewayDescribeCommand())
	cmd.AddCommand(c.newGatewayListCommand())
	cmd.AddCommand(c.newGatewayUpdateCommand())

	return cmd
}

func addGatewayTypeFlag(cmd *cobra.Command) {
	cmd.Flags().String("type", "", fmt.Sprintf("Specify the gateway type as %s.", utils.ArrayToCommaDelimitedString(createGatewayTypes, "or")))
	pcmd.RegisterFlagCompletionFunc(cmd, "type", func(_ *cobra.Command, _ []string) []string { return createGatewayTypes })
}

func (c *command) addRegionFlagGateway(cmd *cobra.Command, command *pcmd.AuthenticatedCLICommand) {
	cmd.Flags().String("region", "", "AWS or Azure region of the gateway.")
	pcmd.RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		regions, err := network.ListRegions(command.Client, cloud)
		if err != nil {
			return nil
		}

		suggestions := make([]string, len(regions))
		for i, region := range regions {
			suggestions[i] = region.RegionId
		}
		return suggestions
	})
}

func (c *command) validGatewayArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validGatewayArgsMultiple(cmd, args)
}

func (c *command) validGatewayArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	return autocompleteGateways(c.V2Client, environmentId)
}

func autocompleteGateways(client *ccloudv2.Client, environmentId string) []string {
	gateways, err := client.ListGateways(environmentId, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(gateways))
	for i, gateway := range gateways {
		suggestions[i] = fmt.Sprintf("%s\t%s", gateway.GetId(), gateway.Spec.GetDisplayName())
	}
	return suggestions
}

func getGatewayCloud(gateway networkinggatewayv1.NetworkingV1Gateway) string {
	cloud := gateway.Status.GetCloudGateway()

	if cloud.NetworkingV1AwsEgressPrivateLinkGatewayStatus != nil || cloud.NetworkingV1AwsIngressPrivateLinkGatewayStatus != nil || cloud.NetworkingV1AwsPrivateNetworkInterfaceGatewayStatus != nil {
		return pcloud.Aws
	}

	if cloud.NetworkingV1AzureEgressPrivateLinkGatewayStatus != nil {
		return pcloud.Azure
	}

	if cloud.NetworkingV1GcpEgressPrivateServiceConnectGatewayStatus != nil {
		return pcloud.Gcp
	}

	if cloud.NetworkingV1GcpPeeringGatewayStatus != nil {
		return pcloud.Gcp
	}

	return ""
}

func getGatewayType(gateway networkinggatewayv1.NetworkingV1Gateway) (string, error) {
	config := gateway.Spec.GetConfig()

	if config.NetworkingV1AwsPrivateNetworkInterfaceGatewaySpec != nil {
		return awsPrivateNetworkInterface, nil
	}

	if config.NetworkingV1AwsEgressPrivateLinkGatewaySpec != nil {
		return awsEgressPrivateLink, nil
	}

	if config.NetworkingV1AwsIngressPrivateLinkGatewaySpec != nil {
		return awsIngressPrivateLink, nil
	}

	if config.NetworkingV1AzureEgressPrivateLinkGatewaySpec != nil {
		return azureEgressPrivateLink, nil
	}

	if config.NetworkingV1AwsPeeringGatewaySpec != nil {
		return awsPeering, nil
	}

	if config.NetworkingV1AzurePeeringGatewaySpec != nil {
		return azurePeering, nil
	}

	if config.NetworkingV1GcpPeeringGatewaySpec != nil {
		return gcpPeering, nil
	}

	if config.NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec != nil {
		return gcpEgressPrivateServiceConnect, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "config")
}

func printGatewayTable(cmd *cobra.Command, gateway networkinggatewayv1.NetworkingV1Gateway) error {
	if gateway.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if gateway.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	gatewayType, err := getGatewayType(gateway)
	if err != nil {
		return err
	}

	out := &gatewayOut{
		Id:           gateway.GetId(),
		Name:         gateway.Spec.GetDisplayName(),
		Environment:  gateway.Spec.Environment.GetId(),
		Type:         gatewayType,
		Phase:        gateway.Status.GetPhase(),
		ErrorMessage: gateway.Status.GetErrorMessage(),
	}

	if gatewayType == awsEgressPrivateLink {
		out.Region = gateway.Spec.Config.NetworkingV1AwsEgressPrivateLinkGatewaySpec.GetRegion()
	}
	if gatewayType == awsIngressPrivateLink {
		out.Region = gateway.Spec.Config.NetworkingV1AwsIngressPrivateLinkGatewaySpec.GetRegion()
	}
	if gatewayType == awsPeering {
		out.Region = gateway.Spec.Config.NetworkingV1AwsPeeringGatewaySpec.GetRegion()
	}
	if gatewayType == azureEgressPrivateLink {
		out.Region = gateway.Spec.Config.NetworkingV1AzureEgressPrivateLinkGatewaySpec.GetRegion()
	}
	if gatewayType == azurePeering {
		out.Region = gateway.Spec.Config.NetworkingV1AzurePeeringGatewaySpec.GetRegion()
	}
	if gatewayType == awsPrivateNetworkInterface {
		out.Region = gateway.Spec.Config.NetworkingV1AwsPrivateNetworkInterfaceGatewaySpec.GetRegion()
		out.Zones = gateway.Spec.Config.NetworkingV1AwsPrivateNetworkInterfaceGatewaySpec.GetZones()
	}
	if gatewayType == gcpEgressPrivateServiceConnect {
		out.Region = gateway.Spec.Config.NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec.GetRegion()
	}
	if gatewayType == gcpPeering {
		out.Region = gateway.Spec.Config.NetworkingV1GcpPeeringGatewaySpec.GetRegion()
	}

	switch getGatewayCloud(gateway) {
	case pcloud.Aws:
		if gatewayType == awsEgressPrivateLink {
			out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsEgressPrivateLinkGatewayStatus.GetPrincipalArn()
		} else if gatewayType == awsIngressPrivateLink {
			out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsIngressPrivateLinkGatewayStatus.GetPrincipalArn()
		} else if gatewayType == awsPrivateNetworkInterface {
			out.Account = gateway.Status.CloudGateway.NetworkingV1AwsPrivateNetworkInterfaceGatewayStatus.GetAccount()
		}
	case pcloud.Azure:
		out.AzureSubscription = gateway.Status.CloudGateway.NetworkingV1AzureEgressPrivateLinkGatewayStatus.GetSubscription()
	case pcloud.Gcp:
		out.GcpProject = gateway.Status.CloudGateway.NetworkingV1GcpEgressPrivateServiceConnectGatewayStatus.GetProject()
		out.GcpIamPrincipal = gateway.Status.CloudGateway.NetworkingV1GcpPeeringGatewayStatus.GetIamPrincipal()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
