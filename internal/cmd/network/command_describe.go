package network

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"

	networking "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"
)

type out struct {
	Id                    string                                 `human:"ID" serialized:"id"`
	EnvironmentId         string                                 `human:"Environment ID" serialized:"environment_id"`
	Name                  string                                 `human:"Name" serialized:"name"`
	Cloud                 string                                 `human:"Cloud" serialized:"cloud"`
	Region                string                                 `human:"Region" serialized:"region"`
	Cidr                  string                                 `human:"CIDR" serialized:"cidr"`
	Zones                 []string                               `human:"Zones" serialized:"zone"`
	DnsResolution         string                                 `human:"DNS Resolution" serialized:"dns_resolution"`
	Phase                 string                                 `human:"Phase" serialized:"phase"`
	ActiveConnectionTypes networking.NetworkingV1ConnectionTypes `human:"Active Connection Types" serialized:"active_connection_types"`
}

func (c *command) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <id>",
		Short: "Desribe a network.",
		Args:  cobra.ExactArgs(1),
		// ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE: c.describe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe Confluent network "n-abcde1".`,
				Code: `confluent network describe n-abcde1`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) describe(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	network, err := c.V2Client.GetNetwork(environmentId, args[0])
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&out{
		Id:                    network.GetId(),
		EnvironmentId:         network.Spec.Environment.GetId(),
		Name:                  network.Spec.GetDisplayName(),
		Cloud:                 network.Spec.GetCloud(),
		Region:                network.Spec.GetRegion(),
		Cidr:                  network.Spec.GetCidr(),
		Zones:                 network.Spec.GetZones(),
		DnsResolution:         network.Spec.DnsConfig.GetResolution(),
		Phase:                 network.Status.GetPhase(),
		ActiveConnectionTypes: network.Status.GetActiveConnectionTypes(),
	})
	return table.Print()
}
