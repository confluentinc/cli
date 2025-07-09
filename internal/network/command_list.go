package network

import (
	"fmt"
	"sort"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List networks.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().StringSlice("name", nil, "A comma-separated list of network names.")
	pcmd.AddListCloudFlag(cmd)
	c.addListRegionFlagNetwork(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().StringSlice("cidr", nil, "A comma-separated list of /16 IPv4 CIDR blocks.")
	addPhaseFlag(cmd, resource.Network)
	addConnectionTypesFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) list(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	name, err := cmd.Flags().GetStringSlice("name")
	if err != nil {
		return err
	}

	cloud, err := cmd.Flags().GetStringSlice("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetStringSlice("region")
	if err != nil {
		return err
	}

	cidr, err := cmd.Flags().GetStringSlice("cidr")
	if err != nil {
		return err
	}

	phase, err := cmd.Flags().GetStringSlice("phase")
	if err != nil {
		return err
	}

	connectionType, err := cmd.Flags().GetStringSlice("connection-types")
	if err != nil {
		return err
	}

	cloud, phase, connectionType = toUpper(cloud), toUpper(phase), toUpper(connectionType)

	networks, err := getNetworks(c.V2Client, environmentId, name, cloud, region, cidr, phase, connectionType)
	if err != nil {
		return err
	}

	// Sort networks by Cloud, then Region, then CreatedAt ASC.
	sort.Slice(networks, func(i, j int) bool {
		if networks[i].Spec.GetCloud() != networks[j].Spec.GetCloud() {
			return networks[i].Spec.GetCloud() < networks[j].Spec.GetCloud()
		}
		if networks[i].Spec.GetRegion() != networks[j].Spec.GetRegion() {
			return networks[i].Spec.GetRegion() < networks[j].Spec.GetRegion()
		}

		return networks[i].Metadata.GetCreatedAt().Before(networks[j].Metadata.GetCreatedAt())
	})

	list := output.NewList(cmd)
	for _, network := range networks {
		if network.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if network.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		zoneInfoStr, err := formatZoneInfoItems(network.Spec.GetZonesInfo())
		if err != nil {
			return err
		}

		list.Add(&out{
			Id:                    network.GetId(),
			Name:                  network.Spec.GetDisplayName(),
			Environment:           network.Spec.Environment.GetId(),
			Gateway:               network.Spec.GetGateway().Id,
			Cloud:                 network.Spec.GetCloud(),
			Region:                network.Spec.GetRegion(),
			Cidr:                  network.Spec.GetCidr(),
			Zones:                 network.Spec.GetZones(),
			ZoneInfo:              zoneInfoStr,
			DnsResolution:         network.Spec.DnsConfig.GetResolution(),
			Phase:                 network.Status.GetPhase(),
			ActiveConnectionTypes: network.Status.GetActiveConnectionTypes(),
		})
	}
	list.Sort(false)
	list.Filter([]string{"Id", "Name", "Gateway", "Cloud", "Region", "Cidr", "Zones", "DnsResolution", "Phase", "ActiveConnectionTypes", "ZoneInfo"})
	return list.Print()
}
