package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingaccesspointv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-access-point/v1"

	"github.com/confluentinc/cli/v3/pkg/ccloudv2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type recordOut struct {
	Id                     string `human:"ID" serialized:"id"`
	Name                   string `human:"Name,omitempty" serialized:"name,omitempty"`
	Domain                 string `human:"Domain" serialized:"domain"`
	PrivateLinkAccessPoint string `human:"Private Link Access Point" serialized:"private_link_access_point"`
	Environment            string `human:"Environment" serialized:"environment"`
	Gateway                string `human:"Gateway" serialized:"gateway"`
	Phase                  string `human:"Phase" serialized:"phase"`
}

func (c *command) newDnsRecordCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage DNS records.",
	}

	cmd.AddCommand(c.newDnsRecordCreateCommand())
	cmd.AddCommand(c.newDnsRecordDeleteCommand())
	cmd.AddCommand(c.newDnsRecordDescribeCommand())
	cmd.AddCommand(c.newDnsRecordListCommand())
	cmd.AddCommand(c.newDnsRecordUpdateCommand())

	return cmd
}

func (c *command) addPrivateLinkAccessPointFlag(cmd *cobra.Command) {
	cmd.Flags().String("private-link-access-point", "", "ID of associated PrivateLink Access Point.")
	pcmd.RegisterFlagCompletionFunc(cmd, "private-link-access-point", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		environmentId, err := c.Context.EnvironmentId()
		if err != nil {
			return nil
		}

		return autocompleteAccessPoints(c.V2Client, environmentId)
	})
}

func autocompleteAccessPoints(client *ccloudv2.Client, environmentId string) []string {
	accessPoints, err := client.ListAccessPoints(environmentId, nil)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(accessPoints))
	for i, accessPoint := range accessPoints {
		suggestions[i] = fmt.Sprintf("%s\t%s", accessPoint.GetId(), accessPoint.Spec.GetDisplayName())
	}
	return suggestions
}

func (c *command) validDnsRecordArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}
	return c.validDnsRecordArgsMultiple(cmd, args)
}

func (c *command) validDnsRecordArgsMultiple(cmd *cobra.Command, args []string) []string {
	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return c.autocompleteDnsRecords()
}

func (c *command) autocompleteDnsRecords() []string {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil
	}

	records, err := c.V2Client.ListDnsRecords(environmentId, ccloudv2.DnsRecordListParameters{})
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(records))
	for i, record := range records {
		suggestions[i] = fmt.Sprintf("%s\t%s", record.GetId(), record.Spec.GetDisplayName())
	}
	return suggestions
}

func printDnsRecordTable(cmd *cobra.Command, record networkingaccesspointv1.NetworkingV1DnsRecord) error {
	if record.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if record.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)

	table.Add(&recordOut{
		Id:                     record.GetId(),
		Name:                   record.Spec.GetDisplayName(),
		Domain:                 record.Spec.GetDomain(),
		PrivateLinkAccessPoint: record.Spec.Config.NetworkingV1PrivateLinkAccessPoint.GetResourceId(),
		Gateway:                record.Spec.Gateway.GetId(),
		Environment:            record.Spec.Environment.GetId(),
		Phase:                  record.Status.GetPhase(),
	})

	return table.Print()
}
