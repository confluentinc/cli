package iam

import (
	"slices"
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type ipFilterCommand struct {
	*pcmd.AuthenticatedCLICommand
}

type ipFilterHumanOut struct {
	ID            string `human:"ID" serialized:"id"`
	Name          string `human:"Name" serialized:"name"`
	ResourceGroup string `human:"Resource group" serialized:"resource_group"`
	IpGroups      string `human:"IP groups" serialized:"ip_groups"`
}

type ipFilterSerializedOut struct {
	ID            string   `human:"ID" serialized:"id"`
	Name          string   `human:"Name" serialized:"name"`
	ResourceGroup string   `human:"Resource group" serialized:"resource_group"`
	IpGroups      []string `human:"IP groups" serialized:"ip_groups"`
}

func newIpFilterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "ip-filter",
		Short:       "Manage IP filters.",
		Long:        "Manage IP filters and their permissions.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
	}

	c := &ipFilterCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())
	cmd.AddCommand(c.newDeleteCommand())
	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func printIpFilter(cmd *cobra.Command, ipFilter iamv2.IamV2IpFilter) error {
	ipGroups := ipFilter.GetIpGroups()
	ipGroupIds := make([]string, len(ipGroups))
	for i, group := range ipGroups {
		ipGroupIds[i] = group.GetId()
	}
	slices.Sort(ipGroupIds)
	table := output.NewTable(cmd)

	if output.GetFormat(cmd) == output.Human {
		table.Add(&ipFilterHumanOut{
			ID:            ipFilter.GetId(),
			Name:          ipFilter.GetFilterName(),
			ResourceGroup: ipFilter.GetResourceGroup(),
			IpGroups:      strings.Join(ipGroupIds, ", "),
		})
	} else {
		table.Add(&ipFilterSerializedOut{
			ID:            ipFilter.GetId(),
			Name:          ipFilter.GetFilterName(),
			ResourceGroup: ipFilter.GetResourceGroup(),
			IpGroups:      ipGroupIds,
		})
	}
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
