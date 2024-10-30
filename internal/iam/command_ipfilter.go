package iam

import (
	"slices"

	"github.com/confluentinc/cli/v4/pkg/config"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type ipFilterCommand struct {
	*pcmd.AuthenticatedCLICommand
	cfg *config.Config
}

type ipFilterOut struct {
	ID              string   `human:"ID" serialized:"id"`
	Name            string   `human:"Name" serialized:"name"`
	ResourceGroup   string   `human:"Resource Group" serialized:"resource_group"`
	IpGroups        []string `human:"IP Groups" serialized:"ip_groups"`
	OperationGroups []string `human:"Operation Groups" serialized:"operation_groups,omitempty"`
	ResourceScope   string   `human:"Resource Scope" serialized:"resource_scope"`
}

func newIpFilterCommand(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ip-filter",
		Short:       "Manage IP filters.",
		Long:        "Manage IP filters and their permissions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ipFilterCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner), cfg}

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
	table.Add(&ipFilterOut{
		ID:              ipFilter.GetId(),
		Name:            ipFilter.GetFilterName(),
		ResourceGroup:   ipFilter.GetResourceGroup(),
		IpGroups:        ipGroupIds,
		ResourceScope:   ipFilter.GetResourceScope(),
		OperationGroups: ipFilter.GetOperationGroups(),
	})
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
