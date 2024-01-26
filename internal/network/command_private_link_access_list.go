package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

type listPrivateLinkAccessOut struct {
	Id           string `human:"ID" serialized:"id"`
	Name         string `human:"Name" serialized:"name"`
	Network      string `human:"Network" serialized:"network"`
	Cloud        string `human:"Cloud" serialized:"cloud"`
	CloudAccount string `human:"Cloud Account,omitempty" serialized:"cloud_account,omitempty"`
	Phase        string `human:"Phase" serialized:"phase"`
}

func (c *command) newPrivateLinkAccessListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List private link accesses.",
		Args:  cobra.NoArgs,
		RunE:  c.privateLinkAccessList,
	}

	cmd.Flags().StringSlice("name", nil, "A comma-separated list of private link access names.")
	addListNetworkFlag(cmd, c.AuthenticatedCLICommand)
	addPhaseFlag(cmd, resource.PrivateLinkAccess)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) privateLinkAccessList(cmd *cobra.Command, _ []string) error {
	name, err := cmd.Flags().GetStringSlice("name")
	if err != nil {
		return err
	}

	network, err := cmd.Flags().GetStringSlice("network")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	phase = toUpper(phase)

	accesses, err := c.getPrivateLinkAccesses(name, network, phase)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, access := range accesses {
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

		out := &listPrivateLinkAccessOut{
			Id:      access.GetId(),
			Name:    access.Spec.GetDisplayName(),
			Network: access.Spec.Network.GetId(),
			Cloud:   cloud,
			Phase:   access.Status.GetPhase(),
		}

		switch cloud {
		case CloudAws:
			out.CloudAccount = access.Spec.Cloud.NetworkingV1AwsPrivateLinkAccess.GetAccount()
		case CloudGcp:
			out.CloudAccount = access.Spec.Cloud.NetworkingV1GcpPrivateServiceConnectAccess.GetProject()
		case CloudAzure:
			out.CloudAccount = access.Spec.Cloud.NetworkingV1AzurePrivateLinkAccess.GetSubscription()
		}

		list.Add(out)
	}
	return list.Print()
}
