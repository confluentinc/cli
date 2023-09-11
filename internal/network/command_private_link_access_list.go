package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type listPrivateLinkAccessOut struct {
	Id           string `human:"ID" serialized:"id"`
	Name         string `human:"Name" serialized:"name"`
	NetworkId    string `human:"Network ID" serialized:"network_id"`
	Cloud        string `human:"Cloud" serialized:"cloud"`
	CloudAccount string `human:"Cloud Account,omitempty" serialized:"cloud_account,omitempty"`
	Phase        string `human:"Phase" serialized:"phase"`
}

func (c *privateLinkAccessCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List private link accesses.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *privateLinkAccessCommand) list(cmd *cobra.Command, _ []string) error {
	accesses, err := c.getPrivateLinkAccesses()
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

		cloud, err := c.getCloud(access)
		if err != nil {
			return err
		}

		out := &listPrivateLinkAccessOut{
			Id:        access.GetId(),
			Name:      access.Spec.GetDisplayName(),
			NetworkId: access.Spec.Network.GetId(),
			Cloud:     cloud,
			Phase:     access.Status.GetPhase(),
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
