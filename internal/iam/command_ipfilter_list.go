package iam

import (
	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"
	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/spf13/cobra"
	"strconv"
)

func (c *ipFilterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP filters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	cmd.Flags().String("environment", "", "Name of the environment for which this filter applies. By default will apply to the org only.")
	cmd.Flags().Bool("include-only-org-scope-filters", false, "Include only org scoped filters as part of the list result.")
	cmd.Flags().Bool("include-parent-scope", true, "If an environment is specified, include org scoped filters in the List Filters response.")

	cmd.MarkFlagsMutuallyExclusive("environment", "include-only-org-scope-filters")
	cmd.MarkFlagsMutuallyExclusive("include-parent-scope", "include-only-org-scope-filters")

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
		resourceScope = crnBase + organizationStr + orgId + environmentStr + environment
	}

	includeOnlyOrgScopeFilters, err := cmd.Flags().GetBool("include-only-org-scope-filters")
	if err != nil {
		return err
	}
	orgOnly := ""
	if cmd.Flags().Changed("include-only-org-scope-filters") {
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
