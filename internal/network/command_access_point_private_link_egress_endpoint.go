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

func (c *accessPointCommand) getEgressEndpoints() ([]networkingaccesspointv1.NetworkingV1AccessPoint, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListAccessPoints(environmentId)
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
	accesses, err := c.getEgressEndpoints()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(accesses))
	for i, access := range accesses {
		suggestions[i] = fmt.Sprintf("%s\t%s", access.GetId(), access.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateLinkEgressEndpointTable(cmd *cobra.Command, accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) error {
	if accessPoint.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if accessPoint.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	out := &accessPointOut{
		Id:          accessPoint.GetId(),
		Name:        accessPoint.Spec.GetDisplayName(),
		Gateway:     accessPoint.Spec.Gateway.GetId(),
		Environment: accessPoint.Spec.Environment.GetId(),
		Phase:       accessPoint.Status.GetPhase(),
	}

	if accessPoint.Status.Config != nil && accessPoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus != nil {
		out.AwsVpcEndpointService = accessPoint.Status.Config.NetworkingV1AwsEgressPrivateLinkEndpointStatus.GetVpcEndpointId()
	}

	if accessPoint.Status.Config != nil && accessPoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus != nil {
		out.AzurePrivateLinkPrivateEndpoint = accessPoint.Status.Config.NetworkingV1AzureEgressPrivateLinkEndpointStatus.GetPrivateEndpointResourceId()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
