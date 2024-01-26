package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type privateLinkAccessOut struct {
	Id                string `human:"ID" serialized:"id"`
	Name              string `human:"Name,omitempty" serialized:"name,omitempty"`
	Network           string `human:"Network" serialized:"network"`
	Cloud             string `human:"Cloud" serialized:"cloud"`
	AwsAccount        string `human:"AWS Account,omitempty" serialized:"aws_account,omitempty"`
	GcpProject        string `human:"GCP Project,omitempty" serialized:"gcp_project,omitempty"`
	AzureSubscription string `human:"Azure Subscription,omitempty" serialized:"azure_subscription,omitempty"`
	Phase             string `human:"Phase" serialized:"phase"`
}

func (c *command) newPrivateLinkAccessCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "access",
		Short: "Manage private link accesses.",
		Args:  cobra.NoArgs,
	}

	cmd.AddCommand(c.newPrivateLinkAccessCreateCommand())
	cmd.AddCommand(c.newPrivateLinkAccessDeleteCommand())
	cmd.AddCommand(c.newPrivateLinkAccessDescribeCommand())
	cmd.AddCommand(c.newPrivateLinkAccessListCommand())
	cmd.AddCommand(c.newPrivateLinkAccessUpdateCommand())

	return cmd
}

func (c *command) getPrivateLinkAccesses(name, network, phase []string) ([]networkingv1.NetworkingV1PrivateLinkAccess, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListPrivateLinkAccesses(environmentId, name, network, phase)
}

func (c *command) validPrivateLinkAccessArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validPrivateLinkAccessArgsMultiple(cmd, args)
}

func (c *command) validPrivateLinkAccessArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompletePrivateLinkAccesses()
}

func (c *command) autocompletePrivateLinkAccesses() []string {
	accesses, err := c.getPrivateLinkAccesses(nil, nil, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(accesses))
	for i, access := range accesses {
		suggestions[i] = fmt.Sprintf("%s\t%s", access.GetId(), access.Spec.GetDisplayName())
	}
	return suggestions
}

func getPrivateLinkAccessCloud(access networkingv1.NetworkingV1PrivateLinkAccess) (string, error) {
	cloud := access.Spec.GetCloud()

	if cloud.NetworkingV1AwsPrivateLinkAccess != nil {
		return CloudAws, nil
	} else if cloud.NetworkingV1GcpPrivateServiceConnectAccess != nil {
		return CloudGcp, nil
	} else if cloud.NetworkingV1AzurePrivateLinkAccess != nil {
		return CloudAzure, nil
	}

	return "", fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "cloud")
}

func printPrivateLinkAccessTable(cmd *cobra.Command, access networkingv1.NetworkingV1PrivateLinkAccess) error {
	if access.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if access.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	cloud, err := getPrivateLinkAccessCloud(access)
	if err != nil {
		return err
	}

	out := &privateLinkAccessOut{
		Id:      access.GetId(),
		Name:    access.Spec.GetDisplayName(),
		Network: access.Spec.Network.GetId(),
		Cloud:   cloud,
		Phase:   access.Status.GetPhase(),
	}

	switch cloud {
	case CloudAws:
		out.AwsAccount = access.Spec.Cloud.NetworkingV1AwsPrivateLinkAccess.GetAccount()
	case CloudGcp:
		out.GcpProject = access.Spec.Cloud.NetworkingV1GcpPrivateServiceConnectAccess.GetProject()
	case CloudAzure:
		out.AzureSubscription = access.Spec.Cloud.NetworkingV1AzurePrivateLinkAccess.GetSubscription()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.Print()
}
