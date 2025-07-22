package iam

import (
	"fmt"
	"sort"
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
	cmd.Flags().String("environment", "", "Identifier of the environment for which this filter applies. Without this flag, applies only to the organization.")
	cmd.Flags().Bool("include-parent-scopes", false, "Include organization scoped filters when listing filters in an environment.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) list(cmd *cobra.Command, _ []string) error {
	var ipFilters []iamipfilteringv2.IamV2IpFilter
	orgId := c.Context.GetCurrentOrganization()
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	resourceScope := ""
	if environment != "" {
		resourceScope = fmt.Sprintf(resourceScopeStr, orgId, environment)
	}
	includeParentScopes, err := cmd.Flags().GetBool("include-parent-scopes")
	if err != nil {
		return err
	}
	parentScopes := ""
	if cmd.Flags().Changed("include-parent-scopes") {
		parentScopes = strconv.FormatBool(includeParentScopes)
	}
	ipFilters, err = c.V2Client.ListIamIpFilters(resourceScope, parentScopes)
	if err != nil {
		return err
	}
	list := output.NewList(cmd)
	for _, filter := range ipFilters {
		filterOut := ipFilterOut{
			ID:            filter.GetId(),
			Name:          filter.GetFilterName(),
			ResourceGroup: filter.GetResourceGroup(),
			ResourceScope: filter.GetResourceScope(),
			IpGroups:      convertIpGroupObjectsToIpGroupIds(filter),
		}
		if filter.OperationGroups != nil {
			sort.Strings(*filter.OperationGroups)
		}
		filterOut.OperationGroups = filter.GetOperationGroups()
		list.Add(&filterOut)
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
