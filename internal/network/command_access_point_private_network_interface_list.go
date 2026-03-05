package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *accessPointCommand) newPrivateNetworkInterfaceListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List private network interfaces.",
		Args:  cobra.NoArgs,
		RunE:  c.privateNetworkInterfaceList,
	}

	cmd.Flags().StringSlice("names", nil, "A comma-separated list of display names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *accessPointCommand) privateNetworkInterfaceList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	names, err := cmd.Flags().GetStringSlice("names")
	if err != nil {
		return err
	}

	accessPoints, err := c.V2Client.ListAccessPoints(environmentId, names)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, accesspoint := range accessPoints {
		if accesspoint.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if accesspoint.Spec.GetConfig().NetworkingV1AwsPrivateNetworkInterface == nil {
			continue
		}

		out := &privateNetworkInterfaceAccessPointOut{
			Id:                accesspoint.GetId(),
			Name:              accesspoint.Spec.GetDisplayName(),
			Gateway:           accesspoint.Spec.Gateway.GetId(),
			Environment:       accesspoint.Spec.Environment.GetId(),
			NetworkInterfaces: accesspoint.Spec.GetConfig().NetworkingV1AwsPrivateNetworkInterface.GetNetworkInterfaces(),
			Account:           accesspoint.Spec.GetConfig().NetworkingV1AwsPrivateNetworkInterface.GetAccount(),
			EgressRoutes:      accesspoint.Spec.GetConfig().NetworkingV1AwsPrivateNetworkInterface.GetEgressRoutes(),
			Phase:             accesspoint.Status.GetPhase(),
		}

		list.Add(out)
	}

	return list.Print()
}
