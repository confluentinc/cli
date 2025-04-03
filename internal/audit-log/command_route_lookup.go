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

	cmd.Flags().AddFlagSet(pcmd.OnPremMTLSSet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cmd.MarkFlagsRequiredTogether("client-cert-path", "client-key-path")

	return cmd
}

func (c *routeCommand) lookup(cmd *cobra.Command, args []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	resource := args[0]
	opts := &mdsv1.ResolveResourceRouteOpts{Crn: optional.NewString(resource)}
	result, response, err := client.AuditLogConfigurationApi.ResolveResourceRoute(c.createContext(), opts)
	if err != nil {
		return HandleMdsAuditLogApiError(err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
