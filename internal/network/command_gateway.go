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
	awsEgressPrivateLink   = "AwsEgressPrivateLink"
	awsPeering             = "AwsPeering"
	azureEgressPrivateLink = "AzureEgressPrivateLink"
	azurePeering           = "AzurePeering"
)

var (
	createGatewayTypes = []string{"egress-privatelink"}
	listGatewayTypes   = []string{"aws-egress-privatelink", "azure-egress-privatelink"}
	gatewayTypeMap     = map[string]string{
		"aws-egress-privatelink":   awsEgressPrivateLink,
		"azure-egress-privatelink": azureEgressPrivateLink,
	}
)

type gatewayOut struct {
	Id                string `human:"ID" serialized:"id"`
	Name              string `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment       string `human:"Environment" serialized:"environment"`
	Region            string `human:"Region,omitempty" serialized:"region,omitempty"`
	Type              string `human:"Type,omitempty" serialized:"type,omitempty"`
	AwsPrincipalArn   string `human:"AWS Principal ARN,omitempty" serialized:"aws_principal_arn,omitempty"`
	AzureSubscription string `human:"Azure Subscription,omitempty" serialized:"azure_subscription,omitempty"`
	Phase             string `human:"Phase" serialized:"phase"`
	ErrorMessage      string `human:"Error Message,omitempty" serialized:"error_message,omitempty"`
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

	if cloud.NetworkingV1AwsEgressPrivateLinkGatewayStatus != nil {
		return pcloud.Aws
	}

	if cloud.NetworkingV1AzureEgressPrivateLinkGatewayStatus != nil {
		return pcloud.Azure
	}

	return ""
}

func printGatewayTable(cmd *cobra.Command, gateway networkinggatewayv1.NetworkingV1Gateway) error {
	if gateway.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if gateway.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	out := &gatewayOut{
		Id:           gateway.GetId(),
		Name:         gateway.Spec.GetDisplayName(),
		Environment:  gateway.Spec.Environment.GetId(),
		Phase:        gateway.Status.GetPhase(),
		ErrorMessage: gateway.Status.GetErrorMessage(),
	}

	if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AwsEgressPrivateLinkGatewaySpec != nil {
		out.Region = gateway.Spec.Config.NetworkingV1AwsEgressPrivateLinkGatewaySpec.GetRegion()
		out.Type = awsEgressPrivateLink
	}
	if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AwsPeeringGatewaySpec != nil {
		out.Region = gateway.Spec.Config.NetworkingV1AwsPeeringGatewaySpec.GetRegion()
		out.Type = awsPeering
	}
	if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AzureEgressPrivateLinkGatewaySpec != nil {
		out.Region = gateway.Spec.Config.NetworkingV1AzureEgressPrivateLinkGatewaySpec.GetRegion()
		out.Type = azureEgressPrivateLink
	}
	if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AzurePeeringGatewaySpec != nil {
		out.Region = gateway.Spec.Config.NetworkingV1AzurePeeringGatewaySpec.GetRegion()
		out.Type = azurePeering
	}

	switch getGatewayCloud(gateway) {
	case pcloud.Aws:
		out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsEgressPrivateLinkGatewayStatus.GetPrincipalArn()
	case pcloud.Azure:
		out.AzureSubscription = gateway.Status.CloudGateway.NetworkingV1AzureEgressPrivateLinkGatewayStatus.GetSubscription()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
