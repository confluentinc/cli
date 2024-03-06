package network

import (
	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

const (
	privateLinkAccessPointKind = "PrivateLinkAccessPoint"
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
				Code: "confluent network dns record create --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com",
			},
			examples.Example{
				Text: "Create a named DNS record.",
				Code: "confluent network dns record create my-dns-record --gateway gw-123456 --private-link-access-point ap-123456 --domain www.example.com",
			},
		),
	}

	c.addPrivateLinkAccessPointFlag(cmd)
	cmd.Flags().String("gateway", "", "Gateway ID.")
	cmd.Flags().String("domain", "", "Fully qualified domain name of the DNS record.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("private-link-access-point"))
	cobra.CheckErr(cmd.MarkFlagRequired("gateway"))
	cobra.CheckErr(cmd.MarkFlagRequired("domain"))

	return cmd
}

func (c *command) dnsRecordCreate(cmd *cobra.Command, args []string) error {
	name := ""
	if len(args) == 1 {
		name = args[0]
	}

	domain, err := cmd.Flags().GetString("domain")
	if err != nil {
		return err
	}

	gateway, err := cmd.Flags().GetString("gateway")
	if err != nil {
		return err
	}

	privateLinkAccessPoint, err := cmd.Flags().GetString("private-link-access-point")
	if err != nil {
		return err
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	createDnsRecord := networkingaccesspointv1.NetworkingV1DnsRecord{
		Spec: &networkingaccesspointv1.NetworkingV1DnsRecordSpec{
			Fqdn: networkingaccesspointv1.PtrString(domain),
			Config: &networkingaccesspointv1.NetworkingV1DnsRecordSpecConfigOneOf{
				NetworkingV1PrivateLinkAccessPoint: &networkingaccesspointv1.NetworkingV1PrivateLinkAccessPoint{
					Kind:       privateLinkAccessPointKind,
					ResourceId: privateLinkAccessPoint,
				},
			},
			Environment: &networkingaccesspointv1.ObjectReference{Id: environmentId},
			Gateway:     &networkingaccesspointv1.EnvScopedObjectReference{Id: gateway},
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
