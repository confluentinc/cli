package network

import (
	"fmt"
	"sort"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List networks.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

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

	networks, err := getNetworks(c.V2Client, environmentId)
	if err != nil {
		return err
	}

	// Sort networks by DisplayName, then Cloud, then Region, then CreatedAt ASC.
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

		zones := network.Spec.GetZones()
		activeConnectionTypes := network.Status.GetActiveConnectionTypes().Items

		if output.GetFormat(cmd) == output.Human {
			list.Add(&humanOut{
				Id:                    network.GetId(),
				Name:                  network.Spec.GetDisplayName(),
				EnvironmentId:         network.Spec.Environment.GetId(),
				Cloud:                 network.Spec.GetCloud(),
				Region:                network.Spec.GetRegion(),
				Cidr:                  network.Spec.GetCidr(),
				Zones:                 strings.Join(zones, ", "),
				DnsResolution:         network.Spec.DnsConfig.GetResolution(),
				Phase:                 network.Status.GetPhase(),
				ActiveConnectionTypes: strings.Join(activeConnectionTypes, ", "),
			})
		} else {
			list.Add(&serializedOut{
				Id:                    network.GetId(),
				Name:                  network.Spec.GetDisplayName(),
				EnvironmentId:         network.Spec.Environment.GetId(),
				Cloud:                 network.Spec.GetCloud(),
				Region:                network.Spec.GetRegion(),
				Cidr:                  network.Spec.GetCidr(),
				Zones:                 zones,
				DnsResolution:         network.Spec.DnsConfig.GetResolution(),
				Phase:                 network.Status.GetPhase(),
				ActiveConnectionTypes: activeConnectionTypes,
			})
		}
	}

	// Disable default sort to use the custom sort above.
	list.Sort(false)
	list.Filter([]string{"Id", "Name", "Cloud", "Region", "Cidr", "Zones", "DnsResolution", "Phase", "ActiveConnectionTypes"})
	return list.Print()
}
