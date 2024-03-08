package network

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDnsRecordListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records.",
		Args:  cobra.NoArgs,
		RunE:  c.dnsRecordList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List DNS records with display names "my-dns-record-1" or "my-dns-record-2.`,
				Code: "confluent network dns record list --names my-dns-record-1,my-dns-record-2",
			},
		),
	}

	cmd.Flags().StringSlice("gateways", nil, "A comma-separated list of gateway IDs.")
	cmd.Flags().StringSlice("names", nil, "A comma-separated list of display names.")
	cmd.Flags().StringSlice("resource-ids", nil, "A comma-separated list of resource IDs.")
	cmd.Flags().StringSlice("domains", nil, "A comma-separated list of fully qualified domain names.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) dnsRecordList(cmd *cobra.Command, _ []string) error {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return err
	}

	gateways, err := cmd.Flags().GetStringSlice("gateways")
	if err != nil {
		return err
	}

	names, err := cmd.Flags().GetStringSlice("names")
	if err != nil {
		return err
	}

	resourceIds, err := cmd.Flags().GetStringSlice("resource-ids")
	if err != nil {
		return err
	}

	domains, err := cmd.Flags().GetStringSlice("domains")
	if err != nil {
		return err
	}

	listParameters := ccloudv2.DnsRecordListParameters{
		Gateways:    gateways,
		Domains:     domains,
		Names:       names,
		ResourceIds: resourceIds,
	}

	records, err := c.V2Client.ListDnsRecords(environmentId, listParameters)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, record := range records {
		if record.Spec == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
		}
		if record.Status == nil {
			return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
		}

		list.Add(&recordOut{
			Id:                     record.GetId(),
			Name:                   record.Spec.GetDisplayName(),
			Domain:                 record.Spec.GetDomain(),
			PrivateLinkAccessPoint: record.Spec.Config.NetworkingV1PrivateLinkAccessPoint.GetResourceId(),
			Gateway:                record.Spec.Gateway.GetId(),
			Environment:            record.Spec.Environment.GetId(),
			Phase:                  record.Status.GetPhase(),
		})
	}

	return list.Print()
}
