package auditlog

import (
	"encoding/json"

	"github.com/antihax/optional"
	"github.com/spf13/cobra"

	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
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
	pcmd.AddMDSOnPremMTLSFlags(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("resource"))

	return cmd
}

func (c *routeCommand) list(cmd *cobra.Command, _ []string) error {
	client, err := c.GetMDSClient(cmd)
	if err != nil {
		return err
	}

	var opts *mdsv1.ListRoutesOpts
	if cmd.Flags().Changed("resource") {
		resource, err := cmd.Flags().GetString("resource")
		if err != nil {
			return err
		}
		opts = &mdsv1.ListRoutesOpts{Q: optional.NewString(resource)}
	} else {
		opts = &mdsv1.ListRoutesOpts{Q: optional.EmptyString()}
	}
	result, response, err := client.AuditLogConfigurationApi.ListRoutes(c.createContext(), opts)
	if err != nil {
		return HandleMdsAuditLogApiError(err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	return enc.Encode(result)
}
