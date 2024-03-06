package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *accessPointCommand) newEgressEndpointCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "egress-endpoint",
		Short: "Manage private link egress endpoints.",
	}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *accessPointCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validArgsMultiple(cmd, args)
}

func (c *accessPointCommand) validArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteEgressEndpoints()
}

func (c *accessPointCommand) autocompleteEgressEndpoints() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	egressEndpoints, err := c.V2Client.ListAccessPoints(environmentId, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(egressEndpoints))
	for i, egressEndpoint := range egressEndpoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", egressEndpoint.GetId(), egressEndpoint.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateLinkEgressEndpointTable(cmd *cobra.Command, egressEndpoint networkingaccesspointv1.NetworkingV1AccessPoint) error {
	if egressEndpoint.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if egressEndpoint.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	out := &accessPointOut{
		Id:          egressEndpoint.GetId(),
		Name:        egressEndpoint.Spec.GetDisplayName(),
		Gateway:     egressEndpoint.Spec.Gateway.GetId(),
		Environment: egressEndpoint.Spec.Environment.GetId(),
		Phase:       egressEndpoint.Status.GetPhase(),
	}

	if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint != nil {
		out.AwsVpcEndpointService = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetVpcEndpointServiceName()
		out.HighAvailability = egressEndpoint.Spec.Config.NetworkingV1AwsEgressPrivateLinkEndpoint.GetEnableHighAvailability()
	}

	if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
		out.AwsVpcEndpoint = egressEndpoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
	}

	if egressEndpoint.Spec.Config != nil && egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint != nil {
		out.AzurePrivateLinkService = egressEndpoint.Spec.Config.NetworkingV1AzureEgressPrivateLinkEndpoint.GetPrivateLinkServiceResourceId()
	}

	if egressEndpoint.Status.Config != nil && egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
		out.AzurePrivateLinkEndpoint = egressEndpoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
