package iam

import (
	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

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

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) list(cmd *cobra.Command, _ []string) error {
	ipFilters, err := c.V2Client.ListIamIpFilters()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, filter := range ipFilters {
		list.Add(&ipFilterOut{
			ID:            filter.GetId(),
			Name:          filter.GetFilterName(),
			ResourceGroup: filter.GetResourceGroup(),
			IpGroups:      convertIpGroupObjectsToIpGroupIds(filter),
		})
	}
	return list.Print()
}

func convertIpGroupObjectsToIpGroupIds(filter iamv2.IamV2IpFilter) []string {
	ipGroups := filter.GetIpGroups()
	ipGroupIds := make([]string, len(ipGroups))
	for i, group := range ipGroups {
		ipGroupIds[i] = group.GetId()
	}
	return ipGroupIds
}
