package network

import (
	"fmt"
	"slices"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type privateNetworkInterfaceAccessPointOut struct {
	Id                string   `human:"ID" serialized:"id"`
	Name              string   `human:"Name,omitempty" serialized:"name,omitempty"`
	Environment       string   `human:"Environment" serialized:"environment"`
	Gateway           string   `human:"Gateway" serialized:"gateway"`
	Phase             string   `human:"Phase" serialized:"phase"`
	NetworkInterfaces []string `human:"Network Interfaces,omitempty" serialized:"network_interfaces,omitempty"`
	Account           string   `human:"Aws Account,omitempty" serialized:"aws_account,omitempty"`
	EgressRoutes      []string `human:"Egress Routes,omitempty" serialized:"egress_routes,omitempty"`
}

func (c *accessPointCommand) newPrivateNetworkInterfaceCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "private-network-interface",
		Short: "Manage access point private network interfaces.",
	}

	cmd.AddCommand(c.newPrivateNetworkInterfaceCreateCommand())
	cmd.AddCommand(c.newPrivateNetworkInterfaceDescribeCommand())
	cmd.AddCommand(c.newPrivateNetworkInterfaceListCommand())
	cmd.AddCommand(c.newPrivateNetworkInterfaceDeleteCommand())
	cmd.AddCommand(c.newPrivateNetworkInterfaceUpdateCommand())

	return cmd
}

func (c *accessPointCommand) validPrivateNetworkInterfaceArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	return c.validPrivateNetworkInterfaceArgsMultiple(cmd, args)
}

func (c *accessPointCommand) validPrivateNetworkInterfaceArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompletePrivateNetworkInterfaces()
}

func (c *accessPointCommand) autocompletePrivateNetworkInterfaces() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	accessPoints, err := c.V2Client.ListAccessPoints(environmentId, nil)
	if err != nil {
		return nil
	}
	privateNetworkInterfaces := slices.DeleteFunc(accessPoints, func(accessPoint networkingaccesspointv1.NetworkingV1AccessPoint) bool {
		return accessPoint.Spec.GetConfig().NetworkingV1AwsPrivateNetworkInterface == nil
	})

	suggestions := make([]string, len(privateNetworkInterfaces))
	for i, privateNetworkInterface := range privateNetworkInterfaces {
		suggestions[i] = fmt.Sprintf("%s\t%s", privateNetworkInterface.GetId(), privateNetworkInterface.Spec.GetDisplayName())
	}
	return suggestions
}

func printPrivateNetworkInterfaceTable(cmd *cobra.Command, privateNetworkInterface networkingaccesspointv1.NetworkingV1AccessPoint) error {
	if privateNetworkInterface.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}

	out := &privateNetworkInterfaceAccessPointOut{
		Id:          privateNetworkInterface.GetId(),
		Name:        privateNetworkInterface.Spec.GetDisplayName(),
		Gateway:     privateNetworkInterface.Spec.Gateway.GetId(),
		Environment: privateNetworkInterface.Spec.Environment.GetId(),
		Phase:       privateNetworkInterface.Status.GetPhase(),
	}

	if privateNetworkInterface.Spec.Config != nil && privateNetworkInterface.Spec.Config.NetworkingV1AwsPrivateNetworkInterface != nil {
		out.NetworkInterfaces = privateNetworkInterface.Spec.Config.NetworkingV1AwsPrivateNetworkInterface.GetNetworkInterfaces()
		out.Account = privateNetworkInterface.Spec.Config.NetworkingV1AwsPrivateNetworkInterface.GetAccount()
		out.EgressRoutes = privateNetworkInterface.Spec.Config.NetworkingV1AwsPrivateNetworkInterface.GetEgressRoutes()
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.PrintWithAutoWrap(false)
}
