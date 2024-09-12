package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcloud "github.com/confluentinc/cli/v3/pkg/cloud"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newGatewayListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List gateways.",
		Args:  cobra.NoArgs,
		RunE:  c.gatewayList,
	}

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

	gateways, err := c.V2Client.ListGateways(environmentId)
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
			Id:          gateway.GetId(),
			Name:        gateway.Spec.GetDisplayName(),
			Environment: gateway.Spec.Environment.GetId(),
			Phase:       gateway.Status.GetPhase(),
		}

		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AwsEgressPrivateLinkGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1AwsEgressPrivateLinkGatewaySpec.GetRegion()
		}
		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AwsPeeringGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1AwsPeeringGatewaySpec.GetRegion()
		}
		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AzureEgressPrivateLinkGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1AzureEgressPrivateLinkGatewaySpec.GetRegion()
		}
		if gateway.Spec.Config != nil && gateway.Spec.Config.NetworkingV1AzurePeeringGatewaySpec != nil {
			out.Region = gateway.Spec.Config.NetworkingV1AzurePeeringGatewaySpec.GetRegion()
		}

		cloud := getGatewayCloud(gateway)

		switch cloud {
		case pcloud.Aws:
			out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsEgressPrivateLinkGatewayStatus.GetPrincipalArn()
		case pcloud.Azure:
			out.AzureSubscription = gateway.Status.CloudGateway.NetworkingV1AzureEgressPrivateLinkGatewayStatus.GetSubscription()
		}

		list.Add(out)
	}
	return list.Print()
}
