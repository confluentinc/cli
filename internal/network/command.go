package network

import (
	"strings"

	"github.com/spf13/cobra"

	networkingv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type humanOut struct {
	Id                    string `human:"ID"`
	EnvironmentId         string `human:"Environment ID"`
	Name                  string `human:"Name"`
	Cloud                 string `human:"Cloud"`
	Region                string `human:"Region"`
	Cidr                  string `human:"CIDR"`
	Zones                 string `human:"Zones"`
	DnsResolution         string `human:"DNS Resolution"`
	Phase                 string `human:"Phase"`
	ActiveConnectionTypes string `human:"Active Connection Types"`
}

type serializedOut struct {
	Id                    string   `serialized:"id"`
	EnvironmentId         string   `serialized:"environment_id"`
	Name                  string   `serialized:"name"`
	Cloud                 string   `serialized:"cloud"`
	Region                string   `serialized:"region"`
	Cidr                  string   `serialized:"cidr"`
	Zones                 []string `serialized:"zones"`
	DnsResolution         string   `serialized:"dns_resolution"`
	Phase                 string   `serialized:"phase"`
	ActiveConnectionTypes []string `serialized:"active_connection_types"`
}

type command struct {
	*pcmd.AuthenticatedCLICommand
}

func New(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "network",
		Short:       "Manage Confluent Cloud networks.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireNonAPIKeyCloudLogin},
	}

	c := &command{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printTable(cmd *cobra.Command, network networkingv1.NetworkingV1Network) error {
	table := output.NewTable(cmd)

	zones := network.Spec.GetZones()
	activeConnectionTypes := network.Status.GetActiveConnectionTypes().Items

	if output.GetFormat(cmd) == output.Human {
		table.Add(&humanOut{
			Id:                    network.GetId(),
			EnvironmentId:         network.Spec.Environment.GetId(),
			Name:                  network.Spec.GetDisplayName(),
			Cloud:                 network.Spec.GetCloud(),
			Region:                network.Spec.GetRegion(),
			Cidr:                  network.Spec.GetCidr(),
			Zones:                 strings.Join(zones, ", "),
			DnsResolution:         network.Spec.DnsConfig.GetResolution(),
			Phase:                 network.Status.GetPhase(),
			ActiveConnectionTypes: strings.Join(activeConnectionTypes, ", "),
		})
	} else {
		table.Add(&serializedOut{
			Id:                    network.GetId(),
			EnvironmentId:         network.Spec.Environment.GetId(),
			Name:                  network.Spec.GetDisplayName(),
			Cloud:                 network.Spec.GetCloud(),
			Region:                network.Spec.GetRegion(),
			Cidr:                  network.Spec.GetCidr(),
			Zones:                 zones,
			DnsResolution:         network.Spec.DnsConfig.GetResolution(),
			Phase:                 network.Status.GetPhase(),
			ActiveConnectionTypes: activeConnectionTypes,
		})
	}

	return table.Print()
}
