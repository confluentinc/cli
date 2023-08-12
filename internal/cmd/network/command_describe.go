package network

import (
	"time"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"

	networking "github.com/confluentinc/ccloud-sdk-go-v2/networking/v1"
)

type out struct {
	Id                       string                                          `human:"Id" serialized:"id"`
	EnvironmentId            string                                          `human:"Environment" serialized:"environment_id"`
	DisplayName              string                                          `human:"Display Name" serialized:"display_name"`
	Cloud                    string                                          `human:"Cloud" serialized:"cloud"`
	Region                   string                                          `human:"Region" serialized:"region"`
	ConnectionTypes          networking.NetworkingV1ConnectionTypes          `human:"ConnectionTypes" serialized:"connection_types"`
	Cidr                     string                                          `human:"Cidr" serialized:"cidr"`
	Zones                    []string                                        `human:"Zones" serialized:"zone"`
	DnsResolution            string                                          `human:"DnsResolution" serialized:"dns_resolution"`
	Phase                    string                                          `human:"Phase" serialized:"phase"`
	SupportedConnectionTypes networking.NetworkingV1SupportedConnectionTypes `human:"SupportedConnectionTypes" serialized:"supported_connection_types"`
	ActiveConnectionTypes    networking.NetworkingV1ConnectionTypes          `human:"ActiveConnectionTypes" serialized:"active_connection_types"`
	ReservedCidr             string                                          `human:"ReservedCidr" serialized:"reserved_cidr"`
	ResourceUrl              string                                          `human:"Resource URL" serialized:"resource_url"`
	ResourceName             string                                          `human:"Resource Name" serialized:"resource_name"`
	CreatedAt                string                                          `human:"Created At" serialized:"created_at"`
	UpdatedAt                string                                          `human:"Updated At" serialized:"updated_at"`
	DeletedAt                string                                          `human:"Deleted At" serialized:"deleted_at"`
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
				Text: `Describe Confluent Network "n-abcde1".`,
				Code: `confluent network describe n-abcde1`,
			},
		),
	}
	pcmd.AddOutputFlag(cmd)
	pcmd.AddClusterFlag(cmd, c.AuthenticatedCLICommand)
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
		Id:                       network.GetId(),
		EnvironmentId:            network.Spec.Environment.GetId(),
		DisplayName:              network.Spec.GetDisplayName(),
		Cloud:                    network.Spec.GetCloud(),
		Region:                   network.Spec.GetRegion(),
		ConnectionTypes:          network.Spec.GetConnectionTypes(),
		Cidr:                     network.Spec.GetCidr(),
		Zones:                    network.Spec.GetZones(),
		DnsResolution:            network.Spec.DnsConfig.GetResolution(),
		ReservedCidr:             network.Spec.GetReservedCidr(),
		Phase:                    network.Status.GetPhase(),
		SupportedConnectionTypes: network.Status.GetSupportedConnectionTypes(),
		ActiveConnectionTypes:    network.Status.GetActiveConnectionTypes(),
		ResourceUrl:              network.Metadata.GetSelf(),
		ResourceName:             network.Metadata.GetResourceName(),
		CreatedAt:                network.Metadata.GetCreatedAt().Format(time.RFC3339),
		UpdatedAt:                network.Metadata.GetUpdatedAt().Format(time.RFC3339),
		DeletedAt:                network.Metadata.GetDeletedAt().Format(time.RFC3339),
	})
	return table.Print()
}
