package iam

import (
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/spf13/cobra"
	"slices"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *ipGroupCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <id>",
		Short: "Update an IP group.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and add a CIDR block to IP group "ipg-12345"":`,
				Code: `confluent iam ip-group update ipg-12345 --name "New Group Name" --add-cidr-blocks "123.234.0.0/16"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Updated name of the IP group.")
	cmd.Flags().StringSlice("add-cidr-blocks", []string{}, "Comma-separated list of CIDR blocks to add.")
	cmd.Flags().StringSlice("remove-cidr-blocks", []string{}, "Comma-separated list of CIDR blocks to remove.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipGroupCommand) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"name",
		"add-cidr-blocks",
		"remove-cidr-blocks",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}

	groupName, err := cmd.Flags().GetString("name")
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

	// Get the current IP group we are going to be updating
	currentIpGroup, err := c.V2Client.GetIamIpGroup(currentIpGroupId)
	// Initialize our new cidr blocks with the existing values
	newCidrBlocks := currentIpGroup.GetCidrBlocks()

	if err != nil {
		return err
	}

	updateIpGroup := iamv2.IamV2IpGroup{Id: &args[0]}
	if groupName != "" {
		updateIpGroup.GroupName = &groupName
	}

	// For each cidr block being added that isn't in the existing slice, append it to the new slice
	if len(addCidrBlocks) > 0 {
		for _, cidrBlock := range addCidrBlocks {
			if !slices.Contains(newCidrBlocks, cidrBlock) {
				newCidrBlocks = append(newCidrBlocks, cidrBlock)
			}
		}
	}
	/*
	 * For each cidr block being removed that is in the existing slice, remove it from the slice.
	 * This is accomplished by recreating the array with every element except for the one being removed
	 */
	if len(removeCidrBlocks) > 0 {
		for _, cidrBlock := range removeCidrBlocks {
			if slices.Contains(newCidrBlocks, cidrBlock) {
				newCidrBlocks = removeIpGroupFromArray(newCidrBlocks, cidrBlock)
			}
		}
	}

	updateIpGroup.CidrBlocks = &newCidrBlocks

	group, err := c.V2Client.UpdateIamIpGroup(updateIpGroup, currentIpGroupId)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanIpGroup(cmd, group)
	}
	return printSerializedIpGroup(cmd, group)
}

func removeIpGroupFromArray(array []string, itemToRemove string) []string {
	for i, element := range array {
		if element == itemToRemove {
			array[i] = array[len(array)-1]
			return array[:len(array)-1]
		}
	}
	return array
}
