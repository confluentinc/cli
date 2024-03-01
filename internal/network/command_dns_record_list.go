package network

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newDnsRecordListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List DNS records.",
		Args:  cobra.NoArgs,
		RunE:  c.dnsRecordList,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) dnsRecordList(cmd *cobra.Command, _ []string) error {
	records, err := c.getDnsRecords()
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
			Id:          record.GetId(),
			Name:        record.Spec.GetDisplayName(),
			Fqdn:        record.Spec.GetFqdn(),
			AccessPoint: record.Spec.Config.NetworkingV1PrivateLinkAccessPoint.GetResourceId(),
			Gateway:     record.Spec.Gateway.GetId(),
			Environment: record.Spec.Environment.GetId(),
			Phase:       record.Status.GetPhase(),
		})
	}

	return list.Print()
}
