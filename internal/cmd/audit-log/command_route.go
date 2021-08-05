package auditlog

import (
	"context"
	"encoding/json"

	"github.com/antihax/optional"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

type routeCommand struct {
	*pcmd.AuthenticatedStateFlagCommand
	prerunner pcmd.PreRunner
}

// NewRouteCommand returns the sub-command object for interacting with audit log route rules.
func NewRouteCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedWithMDSStateFlagCommand(
		&cobra.Command{
			Use:         "route",
			Short:       "Return the audit log route rules.",
			Long:        "Return the routing rules that determine which auditable events are logged, and where.",
			Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		}, prerunner, RouteSubcommandFlags)
	command := &routeCommand{
		AuthenticatedStateFlagCommand: cliCmd,
		prerunner:                     prerunner,
	}
	command.init()
	return command.Command
}

func (c *routeCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List routes matching a resource & sub-resources.",
		Long:  "List the routes that match either the queried resource or its sub-resources.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.list),
	}
	listCmd.Flags().StringP("resource", "r", "", "The Confluent resource name (CRN) that is the subject of the query.")
	check(listCmd.MarkFlagRequired("resource"))
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)

	lookupCmd := &cobra.Command{
		Use:   "lookup <crn>",
		Short: "Return the matching audit-log route rule.",
		Long:  "Return the single route that describes how audit log messages using this CRN would be routed, with all defaults populated.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.lookup),
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
