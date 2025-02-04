package auditlog

import (
	"encoding/json"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
)

func (c *routeCommand) newLookupCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lookup <crn>",
		Short: "Return the matching audit-log route rule.",
		Long:  "Return the single route that describes how audit log messages using this CRN would be routed, with all defaults populated.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.lookup,
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *routeCommand) lookup(_ *cobra.Command, args []string) error {
	resource := args[0]
	opts := &mdsv1.ResolveResourceRouteOpts{Crn: optional.NewString(resource)}
	result, response, err := c.MDSClient.AuditLogConfigurationApi.ResolveResourceRoute(c.createContext(), opts)
	if err != nil {
		return HandleMdsAuditLogApiError(err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
