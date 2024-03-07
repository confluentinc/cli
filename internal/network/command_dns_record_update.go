package network

import (
	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *command) newDnsRecordUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an existing DNS record.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validDnsRecordArgs),
		RunE:              c.dnsRecordUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name of DNS record "dnsrec-123456".`,
				Code: "confluent network dns record update dnsrec-123456 --name my-new-dns-record",
			},
			examples.Example{
				Text: `Update the Privatelink access point of DNS record "dnsrec-123456".`,
				Code: "confluent network dns record update dnsrec-123456 --private-link-access-point ap-123456",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the DNS record.")
	c.addPrivateLinkAccessPointFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "private-link-access-point")

	return cmd
}

func (c *command) dnsRecordUpdate(cmd *cobra.Command, args []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	dnsRecord, err := c.V2Client.GetDnsRecord(environmentId, args[0])
	if err != nil {
		return err
	}

	updateDnsRecord := networkingaccesspointv1.NetworkingV1DnsRecordUpdate{
		Spec: &networkingaccesspointv1.NetworkingV1DnsRecordSpecUpdate{
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Config:      &networkingaccesspointv1.NetworkingV1DnsRecordSpecUpdateConfigOneOf{NetworkingV1PrivateLinkAccessPoint: dnsRecord.Spec.Config.NetworkingV1PrivateLinkAccessPoint},
		},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		updateDnsRecord.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("private-link-access-point") {
		privateLinkAccessPoint, err := cmd.Flags().GetString("private-link-access-point")
		if err != nil {
			return err
		}
		updateDnsRecord.Spec.Config.NetworkingV1PrivateLinkAccessPoint.SetResourceId(privateLinkAccessPoint)
	}

	record, err := c.V2Client.UpdateDnsRecord(args[0], updateDnsRecord)
	if err != nil {
		return err
	}

	return printDnsRecordTable(cmd, record)
}
