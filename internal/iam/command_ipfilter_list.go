package iam

import (
	"fmt"
	"strconv"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *ipFilterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP filters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().String("environment", "", "Name of the environment for which this filter applies. By default will apply to the org only.")
	cmd.Flags().Bool("only-org-filters", false, "Include only org scoped filters as part of the list result.")
	cmd.Flags().Bool("include-parent-scope", true, "If an environment is specified, include org scoped filters in the List Filters response.")

	cmd.MarkFlagsMutuallyExclusive("environment", "only-org-filters")
	cmd.MarkFlagsMutuallyExclusive("include-parent-scope", "only-org-filters")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) list(cmd *cobra.Command, _ []string) error {
	orgId := c.Context.GetCurrentOrganization()
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	resourceScope := ""
	if environment != "" {
		resourceScope = fmt.Sprintf(resourceScopeStr, orgId, environment)
	}

	includeOnlyOrgScopeFilters, err := cmd.Flags().GetBool("only-org-filters")
	if err != nil {
		return err
	}
	orgOnly := ""
	if cmd.Flags().Changed("only-org-filters") {
		orgOnly = strconv.FormatBool(includeOnlyOrgScopeFilters)
	}

	includeParentScope, err := cmd.Flags().GetBool("include-parent-scope")
	if err != nil {
		return err
	}
	parentScope := ""
	if cmd.Flags().Changed("include-parent-scope") {
		parentScope = strconv.FormatBool(includeParentScope)
	}
	ipFilters, err := c.V2Client.ListIamIpFilters(resourceScope, orgOnly, parentScope)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, filter := range ipFilters {
		list.Add(&ipFilterOut{
			ID:              filter.GetId(),
			Name:            filter.GetFilterName(),
			ResourceGroup:   filter.GetResourceGroup(),
			IpGroups:        convertIpGroupObjectsToIpGroupIds(filter),
			OperationGroups: filter.GetOperationGroups(),
			ResourceScope:   filter.GetResourceScope(),
		})
	}
	return list.Print()
}

func convertIpGroupObjectsToIpGroupIds(filter iamipfilteringv2.IamV2IpFilter) []string {
	ipGroups := filter.GetIpGroups()
	ipGroupIds := make([]string, len(ipGroups))
	for i, group := range ipGroups {
		ipGroupIds[i] = group.GetId()
	}
	return ipGroupIds
}
