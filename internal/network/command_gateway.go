package network

import (
	"fmt"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/spf13/cobra"
)

type gatewayOut struct {
	Id                string `human:"ID" serialized:"id"`
	Name              string `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment       string `human:"Environment" serialized:"environment"`
	Region            string `human:"Region,omitempty" serialized:"region,omitempty"`
	AwsPrincipalArn   string `human:"AWS Principal ARN,omitempty" serialized:"aws_principal_arn,omitempty"`
	AzureSubscription string `human:"Azure Subscription,omitempty" serialized:"azure_subscription,omitempty"`
	Phase             string `human:"Phase" serialized:"phase"`
}

func (c *command) newGatewayCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "gateway",
		Short: "Manage gateways.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newGatewayDescribeCommand())
	cmd.AddCommand(c.newGatewayListCommand())

	return cmd
}

func (c *command) getGateways() ([]networkingv1.NetworkingV1Gateway, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListGateways(environmentId)
}

func (c *command) validGatewayArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteGateways()
}

func (c *command) autocompleteGateways() []string {
	gateways, err := c.getGateways()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(gateways))
	for i, gateway := range gateways {
		suggestions[i] = fmt.Sprintf("%s\t%s", gateway.GetId(), gateway.Spec.GetDisplayName())
	}
	return suggestions
}

func getGatewayCloud(gateway networkingv1.NetworkingV1Gateway) (string, error) {
	cloud := gateway.Status.GetCloudGateway()

	if cloud.NetworkingV1AwsEgressPrivateLinkGatewayStatus != nil {
		return CloudAws, nil
	} else if cloud.NetworkingV1AzureEgressPrivateLinkGatewayStatus != nil {
		return CloudAzure, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
}
