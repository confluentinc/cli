package iam

import (
	"slices"
	"sort"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type ipFilterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipFilterOut struct {
	ID              string   `human:"ID" serialized:"id"`
	Name            string   `human:"Name" serialized:"name"`
	ResourceGroup   string   `human:"Resource Group" serialized:"resource_group"`
	IpGroups        []string `human:"IP Groups" serialized:"ip_groups"`
	OperationGroups []string `human:"Operation Groups,omitempty" serialized:"operation_groups,omitempty"`
	ResourceScope   string   `human:"Resource Scope,omitempty" serialized:"resource_scope,omitempty"`
}

func newIpFilterCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ip-filter",
		Short:       "Manage IP filters.",
		Long:        "Manage IP filters and their permissions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ipFilterCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand(cfg))
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand(cfg))
	cmd.AddCommand(c.newUpdateCommand(cfg))

	return cmd
}

func printIpFilter(cmd *cobra.Command, ipFilter iamipfilteringv2.IamV2IpFilter) error {
	ipGroupIds := convertIpGroupsToIds(ipFilter.GetIpGroups())
	slices.Sort(ipGroupIds)
	table := output.NewTable(cmd)
	filterOut := &ipFilterOut{
		ID:            ipFilter.GetId(),
		Name:          ipFilter.GetFilterName(),
		ResourceGroup: ipFilter.GetResourceGroup(),
		ResourceScope: ipFilter.GetResourceScope(),
		IpGroups:      ipGroupIds,
	}
	if ipFilter.OperationGroups != nil {
		sort.Strings(*ipFilter.OperationGroups)
	}
	filterOut.OperationGroups = ipFilter.GetOperationGroups()
	table.Add(filterOut)
	return table.Print()
}

func (c *ipFilterCommand) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return pcmd.AutocompleteIpFilters(c.V2Client)
}

func convertIpGroupsToIds(ipGroups []iamipfilteringv2.GlobalObjectReference) []string {
	ipGroupIds := make([]string, len(ipGroups))
	for i, group := range ipGroups {
		ipGroupIds[i] = group.GetId()
	}
	return ipGroupIds
}
