package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"strings"
)

func (c *ipFilterCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List IP filters.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) list(cmd *cobra.Command, _ []string) error {
	ipFilters, err := c.V2Client.ListIamIpFilters()
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	if output.GetFormat(cmd) == output.Human {
		for _, filter := range ipFilters {
			ipGroupIds := convertIpGroupObjectsToIpGroupIds(filter)
			list.Add(&ipFilterHumanOut{
				ID:            filter.GetId(),
				Name:          filter.GetFilterName(),
				ResourceGroup: filter.GetResourceGroup(),
				IpGroups:      strings.Join(ipGroupIds, ", "),
			})
		}
		return list.Print()
	} else {
		for _, filter := range ipFilters {
			ipGroupIds := convertIpGroupObjectsToIpGroupIds(filter)
			list.Add(&ipFilterSerializedOut{
				ID:            filter.GetId(),
				Name:          filter.GetFilterName(),
				ResourceGroup: filter.GetResourceGroup(),
				IpGroups:      ipGroupIds,
			})
		}
	}
	return list.Print()
}

func convertIpGroupObjectsToIpGroupIds(filter iamv2.IamV2IpFilter) []string {
	var ipGroupIds []string
	for _, group := range filter.GetIpGroups() {
		ipGroupIds = append(ipGroupIds, group.GetId())
	}
	return ipGroupIds
}
