package auditlog

import (
	"encoding/json"
	"github.com/antihax/optional"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/mds-sdk-go"
	"github.com/spf13/cobra"

	"context"
)

type routeCommand struct {
	*cobra.Command
	config *config.Config
	client *mds.APIClient
	ctx    context.Context
}

// NewRouteCommand returns the sub-command object for interacting with audit log route rules.
func NewRouteCommand(config *config.Config, client *mds.APIClient) *cobra.Command {
	cmd := &routeCommand{
		Command: &cobra.Command{
			Use:   "route",
			Short: "Examine audit log route rules.",
			Long:  "Examine routing rules that determine which auditable events are logged, and where.",
		},
		config: config,
		client: client,
		ctx:    context.WithValue(context.Background(), mds.ContextAccessToken, config.AuthToken),
	}

	cmd.init()
	return cmd.Command
}

func (c *routeCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List the routes that could match the queried resource or its sub-resources.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("resource", "", "The confluent resource name (CRN) that is the subject of the query.")
	c.AddCommand(listCmd)

	lookupCmd := &cobra.Command{
		Use:   "lookup <crn>",
		Short: "Returns the single route that describes how audit log messages regarding this CRN would be routed, with all defaults populated.",
		RunE:  c.lookup,
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(lookupCmd)
}

func (c *routeCommand) list(cmd *cobra.Command, args []string) error {
	var opts *mds.ListRoutesOpts
	if c.Flags().Changed("resource") {
		resource, err := c.Flags().GetString("resource")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		opts = &mds.ListRoutesOpts{Q: optional.NewString(resource)}
	} else {
		opts = &mds.ListRoutesOpts{Q: optional.EmptyString()}
	}
	result, _, err := c.client.AuditLogConfigurationApi.ListRoutes(c.ctx, opts)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = json.NewEncoder(c.OutOrStdout()).Encode(result)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}

func (c *routeCommand) lookup(cmd *cobra.Command, args []string) error {
	resource := args[0]
	opts := &mds.ResolveResourceRouteOpts{Crn: optional.NewString(resource)}
	result, _, err := c.client.AuditLogConfigurationApi.ResolveResourceRoute(c.ctx, opts)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	err = json.NewEncoder(c.OutOrStdout()).Encode(result)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	return nil
}
