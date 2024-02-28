package network

import (
	"github.com/spf13/cobra"

	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/networking-outbound-privatelink/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

const (
	privateLinkAccessPoint = "PrivateLinkAccessPoint"
)

func (c *command) newDnsRecordCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create [name]",
		Short: "Create a DNS record.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  c.dnsRecordCreate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Create a DNS record.",
				Code: "confluent network dns record create --gateway gw-123456 --access-point ap-123456 --fqdn www.example.com",
			},
			examples.Example{
				Text: "Create a named DNS record.",
				Code: "confluent network dns record create my-dns-record --gateway gw-123456 --access-point ap-123456 --fqdn www.example.com",
			},
		),
	}

	c.addAccessPointFlag(cmd)
	cmd.Flags().String("gateway", "", "Gateway ID.")
	cmd.Flags().String("fqdn", "", "Fully qualified domain name of the DNS record.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("access-point"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("fqdn"))

	return cmd
}

func (c *command) dnsRecordCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	fqdn, err := cmd.Flags().GetString("fqdn")
	if err != nil {
		return err
	}

	gateway, err := cmd.Flags().GetString("gateway")
	if err != nil {
		return err
	}

	accessPoint, err := cmd.Flags().GetString("access-point")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createDnsRecord := networkingoutboundprivatelinkv1.NetworkingV1DnsRecord{
		Spec: &networkingoutboundprivatelinkv1.NetworkingV1DnsRecordSpec{
			Fqdn: networkingoutboundprivatelinkv1.PtrString(fqdn),
			Config: &networkingoutboundprivatelinkv1.NetworkingV1DnsRecordSpecConfigOneOf{
				NetworkingV1PrivateLinkAccessPoint: &networkingoutboundprivatelinkv1.NetworkingV1PrivateLinkAccessPoint{
					Kind:       privateLinkAccessPoint,
					ResourceId: accessPoint,
				},
			},
			Environment: &networkingoutboundprivatelinkv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingoutboundprivatelinkv1.EnvScopedObjectReference{Id: gateway},
		},
	}

	if name != "" {
		createDnsRecord.Spec.SetDisplayName(name)
	}

	record, err := c.V2Client.CreateDnsRecord(createDnsRecord)
	if err != nil {
		return err
	}

	return printDnsRecordTable(cmd, record)
}
