package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"strings"
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

func newIPFilterCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ip-filter",
		Short: "Manage IP filters",
		Long:  "Manage IP filters and their permissions",
	}

	c := &ipFilterCommand{pcmd.NewAuthenticatedCLICommand(cmd, prerunner)}

	cmd.AddCommand(c.newCreateCommand())

	return cmd
}

func printHumanIPFilter(cmd *cobra.Command, ipFilter iamv2.IamV2IpFilter) error {
	var ipGroupIds []string
	for _, group := range ipFilter.GetIpGroups() {
		ipGroupIds = append(ipGroupIds, group.GetId())
	}
	table := output.NewTable(cmd)
	table.Add(&ipFilterHumanOut{
		ID:            ipFilter.GetId(),
		Name:          ipFilter.GetFilterName(),
		ResourceGroup: ipFilter.GetResourceGroup(),
		IpGroups:      strings.Join(ipGroupIds, ", "),
	})
	return table.Print()
}

func printSerializedIPFilter(cmd *cobra.Command, ipFilter iamv2.IamV2IpFilter) error {
	var ipGroupIds []string
	for _, group := range ipFilter.GetIpGroups() {
		ipGroupIds = append(ipGroupIds, group.GetId())
	}
	table := output.NewTable(cmd)
	table.Add(&ipFilterSerializedOut{
		ID:            ipFilter.GetId(),
		Name:          ipFilter.GetFilterName(),
		ResourceGroup: ipFilter.GetResourceGroup(),
		IpGroups:      ipGroupIds,
	})
	return table.Print()
}
