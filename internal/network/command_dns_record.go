package network

import (
	"fmt"

	"github.com/spf13/cobra"

	networkingoutboundprivatelinkv1 "github.com/confluentinc/ccloud-sdk-go-v2/networking-outbound-privatelink/v1"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type RecordOut struct {
	Id          string `human:"ID" serialized:"id"`
	Name        string `human:"Name,omitempty" serialized:"name,omitempty"`
	Fqdn        string `human:"FQDN" serialized:"fqdn"`
	AccessPoint string `human:"Access Point" serialized:"access_point"`
	Environment string `human:"Environment" serialized:"environment"`
	Gateway     string `human:"Gateway" serialized:"gateway"`
	Phase       string `human:"Phase" serialized:"phase"`
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

func (c *command) addAccessPointFlag(cmd *cobra.Command) {
	cmd.Flags().String("access-point", "", "PrivateLink access point.")
}

func (c *command) getDnsRecords() ([]networkingoutboundprivatelinkv1.NetworkingV1DnsRecord, error) {
	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return nil, err
	}

	return c.V2Client.ListDnsRecords(environmentId)
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
	records, err := c.getDnsRecords()
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(records))
	for i, record := range records {
		suggestions[i] = fmt.Sprintf("%s\t%s", record.GetId(), record.Spec.GetDisplayName())
	}
	return suggestions
}

func printDnsRecordTable(cmd *cobra.Command, record networkingoutboundprivatelinkv1.NetworkingV1DnsRecord) error {
	if record.Spec == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "spec")
	}
	if record.Status == nil {
		return fmt.Errorf(errors.CorruptedNetworkResponseErrorMsg, "status")
	}

	table := output.NewTable(cmd)

	table.Add(&RecordOut{
		Id:          record.GetId(),
		Name:        record.Spec.GetDisplayName(),
		Fqdn:        record.Spec.GetFqdn(),
		AccessPoint: record.Spec.Config.NetworkingV1PrivateLinkAccessPoint.GetResourceId(),
		Gateway:     record.Spec.Gateway.GetId(),
		Environment: record.Spec.Environment.GetId(),
		Phase:       record.Status.GetPhase(),
	})

	return table.Print()
}
