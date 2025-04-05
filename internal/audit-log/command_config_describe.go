package auditlog

import (
	"encoding/json"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *configCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Prints the audit log configuration specification object.",
		Long:  `Prints the audit log configuration specification object, where "specification" refers to the JSON blob that describes audit log routing rules.`,
		Args:  cobra.NoArgs,
		RunE:  c.describe,
	}

	pcmd.AddMDSOnPremMTLSFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *configCommand) describe(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	spec, response, err := client.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return HandleMdsAuditLogApiError(err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(spec)
}
