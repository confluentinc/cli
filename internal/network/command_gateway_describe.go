package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newGatewayDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a gateway.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validGatewayArgs),
		RunE:              c.gatewayDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe gateway "gw-123456".`,
				Code: "confluent network gateway describe gw-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) gatewayDescribe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	gateway, err := c.V2Client.GetGateway(environmentId, args[0])
	if err != nil {
		return err
	}

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
	case CloudAws:
		out.AwsPrincipalArn = gateway.Status.CloudGateway.NetworkingV1AwsEgressPrivateLinkGatewayStatus.GetPrincipalArn()
	case CloudAzure:
		out.AzureSubscription = gateway.Status.CloudGateway.NetworkingV1AzureEgressPrivateLinkGatewayStatus.GetSubscription()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
