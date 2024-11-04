package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcloud "github.com/confluentinc/cli/v4/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

func (c *command) newGatewayListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List gateways.",
		Args:  cobra.NoArgs,
		RunE:  c.gatewayList,
	}

	cmd.Flags().StringSlice("types", nil, fmt.Sprintf("A comma-separated list of gateway types: %s.", utils.ArrayToCommaDelimitedString(listGatewayTypes, "or")))
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) gatewayList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	types, err := cmd.Flags().GetStringSlice("types")
	if err != nil {
		return err
	}
	for i, gatewayType := range types {
		if val, ok := gatewayTypeMap[gatewayType]; ok {
			types[i] = val
		}
	}

	gateways, err := c.V2Client.ListGateways(environmentId, types)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, gateway := range gateways {
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
		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1GcpEgressPrivateServiceConnectGatewaySpec.GetRegion()
			out.Type = gcpEgressPrivateLink
		}
		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1GcpPeeringGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1GcpPeeringGatewaySpec.GetRegion()
			out.Type = gcpPeering
		}

		switch getGatewayCloud(gateway) {
		case pcloud.Aws:
			out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsEgressPrivateLinkGatewayStatus.GetPrincipalArn()
		case pcloud.Azure:
			out.AzureSubscription = gateway.Status.CloudGateway.NetworkingV1AzureEgressPrivateLinkGatewayStatus.GetSubscription()
		case pcloud.Gcp:
			out.GcpProject = gateway.Status.CloudGateway.NetworkingV1GcpEgressPrivateServiceConnectGatewayStatus.GetProject()
		}

		list.Add(out)
	}
	return list.Print()
}
