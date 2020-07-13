package auditlog

import (
	"context"
	"encoding/json"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
)

type routeCommand struct {
	*cmd.AuthenticatedCLICommand
	prerunner cmd.PreRunner
}

// NewRouteCommand returns the sub-command object for interacting with audit log route rules.
func NewRouteCommand(prerunner cmd.PreRunner) *cobra.Command {
	cliCmd := cmd.NewAuthenticatedWithMDSCLICommand(
		&cobra.Command{
			Use:   "route",
			Short: "Examine audit log route rules.",
			Long:  "Examine routing rules that determine which auditable events are logged, and where.",
		}, prerunner)
	command := &routeCommand{
		AuthenticatedCLICommand: cliCmd,
		prerunner:               prerunner,
	}
	command.init()
	return command.Command
}

func (c *routeCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List routes matching a resource & sub-resources.",
		Long:  "List the routes that could match the queried resource or its sub-resources.",
		RunE:  cmd.NewCLIRunE(c.list),
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().StringP("resource", "r", "", "The confluent resource name that is the subject of the query.")
	check(listCmd.MarkFlagRequired("resource"))
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	lookupCmd := &cobra.Command{
		Use:   "lookup <crn>",
		Short: "Returns the matching audit-log route rule.",
		Long:  "Returns the single route that describes how audit log messages regarding this CRN would be routed, with all defaults populated.",
		RunE:  cmd.NewCLIRunE(c.lookup),
		Args:  cobra.ExactArgs(1),
	}
	c.AddCommand(lookupCmd)
}

func (c *routeCommand) createContext() context.Context {
	return context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
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
	if err = enc.Encode(result); err != nil {
		return err
	}
	return nil
}

func (c *routeCommand) lookup(cmd *cobra.Command, args []string) error {
	resource := args[0]
	opts := &mds.ResolveResourceRouteOpts{Crn: optional.NewString(resource)}
	result, response, err := c.MDSClient.AuditLogConfigurationApi.ResolveResourceRoute(c.createContext(), opts)
	if err != nil {
		return HandleMdsAuditLogApiError(cmd, err, response)
	}
	enc := json.NewEncoder(c.OutOrStdout())
	enc.SetIndent("", "  ")
	if err = enc.Encode(result); err != nil {
		return err
	}
	return nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
