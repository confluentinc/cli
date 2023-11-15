package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/spf13/cobra"
	"slices"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *ipGroupCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an IP Group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the Group Name of IP Group "ipg-12345"":`,
				Code: `confluent iam ip-group update ipg-12345 --group-name "New Group Name"`,
			},
		),
	}

	pcmd.AddProviderFlag(cmd, c.AuthenticatedCLICommand)
	cmd.Flags().String("group-name", "", "Name of the IP Group.")
	cmd.Flags().StringSlice("add-cidr-blocks", []string{}, "List of CIDR blocks to add to existing CIDR blocks on the IP Group.")
	cmd.Flags().StringSlice("remove-cidr-blocks", []string{}, "List of CIDR blocks to remove from existing CIDR blocks on the IP Group.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddFilterFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"group-name",
		"add-cidr-blocks",
		"remove-cidr-blocks",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}

	groupName, err := cmd.Flags().GetString("group-name")
	if err != nil {
		return err
	}

	addCidrBlocks, err := cmd.Flags().GetStringSlice("add-cidr-blocks")
	if err != nil {
		return err
	}

	removeCidrBlocks, err := cmd.Flags().GetStringSlice("remove-cidr-blocks")
	if err != nil {
		return err
	}

	currentIpGroupId := args[0]

	// get the current IP group we are going to be updating
	currentIpGroup, err := c.V2Client.GetIamIpGroup(currentIpGroupId)
	// initialize our new cidr blocks with the existing values
	newCidrBlocks := currentIpGroup.GetCidrBlocks()

	if err != nil {
		return err
	}

	updateIpGroup := iamv2.IamV2IpGroup{Id: &args[0]}
	if groupName != "" {
		updateIpGroup.GroupName = &groupName
	}

	// for each cidr block being added that isn't in the existing slice, append it to the new slice
	if len(addCidrBlocks) > 0 {
		for _, cidrBlock := range addCidrBlocks {
			if !slices.Contains(newCidrBlocks, cidrBlock) {
				newCidrBlocks = append(newCidrBlocks, cidrBlock)
			}
		}
	}
	/*
	 * for each cidr block being removed that is in the existing slice, remove it from the slice.
	 * this is accomplished by recreating the array with every element except for the one being removed
	 */
	if len(removeCidrBlocks) > 0 {
		for _, cidrBlock := range removeCidrBlocks {
			if slices.Contains(newCidrBlocks, cidrBlock) {
				newCidrBlocks = removeElementFromArray(newCidrBlocks, cidrBlock)
			}
		}
	}

	updateIpGroup.CidrBlocks = &newCidrBlocks

	group, err := c.V2Client.UpdateIamIpGroup(updateIpGroup, currentIpGroupId)
	if err != nil {
		return err
	}

	return printIPGroup(cmd, group)
}

func removeElementFromArray(array []string, itemToRemove string) []string {
	for i, element := range array {
		if element == itemToRemove {
			array[i] = array[len(array)-1]
			return array[:len(array)-1]
		}
	}
	return array
}
