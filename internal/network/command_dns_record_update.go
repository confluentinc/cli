package network

import (
	"github.com/spf13/cobra"

	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-outbound-privatelink/v1"

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
				Code: "confluent network dns record update dnsrec-123456 --access-point ap-123456",
			},
		),
	}

	cmd.Flags().String("name", "", "Name of the DNS record.")
	c.addAccessPointFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "access-point")

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

	updateDnsRecord := networkingoutboundprivatelinkv1.NetworkingV1DnsRecordUpdate{
		Spec: &networkingoutboundprivatelinkv1.NetworkingV1DnsRecordSpecUpdate{
			Environment: &networkingoutboundprivatelinkv1.ObjectReference{Id: environmentId},
			Config:      &networkingoutboundprivatelinkv1.NetworkingV1DnsRecordSpecUpdateConfigOneOf{NetworkingV1PrivateLinkAccessPoint: dnsRecord.Spec.Config.NetworkingV1PrivateLinkAccessPoint},
		},
	}

	if cmd.Flags().Changed("name") {
		name, err := cmd.Flags().GetString("name")
		if err != nil {
			return err
		}
		updateDnsRecord.Spec.SetDisplayName(name)
	}

	if cmd.Flags().Changed("access-point") {
		accessPoint, err := cmd.Flags().GetString("access-point")
		if err != nil {
			return err
		}
		updateDnsRecord.Spec.Config.NetworkingV1PrivateLinkAccessPoint.SetResourceId(accessPoint)
	}

	record, err := c.V2Client.UpdateDnsRecord(args[0], updateDnsRecord)
	if err != nil {
		return err
	}

	return printDnsRecordTable(cmd, record)
}
