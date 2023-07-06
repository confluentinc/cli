package auditlog

import (
	"encoding/json"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *configCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Prints the audit log configuration specification object.",
		Long:  `Prints the audit log configuration specification object, where "specification" refers to the JSON blob that describes audit log routing rules.`,
		Args:  cobra.NoArgs,
		RunE:  c.describe,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *configCommand) describe(cmd *cobra.Command, _ []string) error {
	spec, response, err := c.MDSClient.AuditLogConfigurationApi.GetConfig(c.createContext())
	if err != nil {
		return HandleMdsAuditLogApiError(cmd, err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(spec)
}
