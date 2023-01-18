package auditlog

import (
	"encoding/json"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go-public/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *routeCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List routes matching a resource & sub-resources.",
		Long:  "List the routes that match either the queried resource or its sub-resources.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().String("resource", "", "The Confluent resource name (CRN) that is the subject of the query.")
	pcmd.AddContextFlag(cmd, c.CLICommand)

	_ = cmd.MarkFlagRequired("resource")

	return cmd
}

func (c *routeCommand) list(cmd *cobra.Command, _ []string) error {
	var opts *mds.ListRoutesOpts
	if cmd.Flags().Changed("resource") {
		resource, err := cmd.Flags().GetString("resource")
		if err != nil {
			return err
		}
		opts = &mds.ListRoutesOpts{Q: optional.NewString(resource)}
	} else {
		opts = &mds.ListRoutesOpts{Q: optional.EmptyString()}
	}
	result, response, err := c.MDSClient.AuditLogConfigurationApi.ListRoutes(c.createContext(), opts)
	if err != nil {
		return HandleMdsAuditLogApiError(cmd, err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
