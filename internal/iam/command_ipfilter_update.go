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

func (c *ipFilterCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an IP filter.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and add an IP group to IP filter "ipf-abcde":`,
				Code: `confluent iam ip-filter update ipf-abcde --name "New Filter Name" --add-ip-groups "ipg-12345"`,
			},
		),
	}

	cmd.Flags().String("name", "", "Updated name of the IP filter.")
	pcmd.AddResourceGroupFlag(cmd)
	cmd.Flags().StringSlice("add-ip-groups", []string{}, "Comma-separated list of IP groups to add.")
	cmd.Flags().StringSlice("remove-ip-groups", []string{}, "Comma-separated list of IP groups to remove.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) update(cmd *cobra.Command, args []string) error {
	flags := []string{
		"name",
		"resource-group",
		"add-ip-groups",
		"remove-ip-groups",
	}
	if err := errors.CheckNoUpdate(cmd.Flags(), flags...); err != nil {
		return err
	}

	filterName, err := cmd.Flags().GetString("name")
	if err != nil {
		return err
	}

	resourceGroup, err := cmd.Flags().GetString("resource-group")
	if err != nil {
		return err
	}

	addIpGroups, err := cmd.Flags().GetStringSlice("add-ip-groups")
	if err != nil {
		return err
	}

	removeIpGroups, err := cmd.Flags().GetStringSlice("remove-ip-groups")
	if err != nil {
		return err
	}

	currentIPFilterId := args[0]

	// Get the current IP group we are going to be updating
	currentIpFilter, err := c.V2Client.GetIamIpFilter(currentIPFilterId)
	// Initialize our new IP groups list with the existing ids
	var newIpGroupIds []string
	for _, group := range currentIpFilter.GetIpGroups() {
		newIpGroupIds = append(newIpGroupIds, group.GetId())
	}

	if err != nil {
		return err
	}

	updateIpFilter := currentIpFilter

	if filterName != "" {
		updateIpFilter.FilterName = &filterName
	}

	if resourceGroup != "" {
		updateIpFilter.ResourceGroup = &resourceGroup
	}

	// For each IP group ID being added that isn't in the existing slice, append it to the new slice
	if len(addIpGroups) > 0 {
		for _, ipGroupId := range addIpGroups {
			if !slices.Contains(newIpGroupIds, ipGroupId) {
				newIpGroupIds = append(newIpGroupIds, ipGroupId)
			}
		}
	}
	/*
	 * For each IP group ID being removed that is in the existing slice, remove it from the slice.
	 * This is accomplished by recreating the array with every element except for the one being removed
	 */
	if len(removeIpGroups) > 0 {
		for _, ipGroupId := range removeIpGroups {
			if slices.Contains(newIpGroupIds, ipGroupId) {
				newIpGroupIds = removeIpFilterFromArray(newIpGroupIds, ipGroupId)
			}
		}
	}

	// Convert the IP group IDs into IP group objects
	var IpGroupIdObjects []iamv2.GlobalObjectReference
	for _, ipGroupId := range newIpGroupIds {
		// The empty string fields will get filled in automatically by the cc-policy-service
		IpGroupIdObjects = append(IpGroupIdObjects, *iamv2.NewGlobalObjectReference(ipGroupId, "", ""))
	}

	updateIpFilter.IpGroups = &IpGroupIdObjects

	filter, err := c.V2Client.UpdateIamIpFilter(updateIpFilter, currentIPFilterId)
	if err != nil {
		return err
	}

	if output.GetFormat(cmd) == output.Human {
		return printHumanIpFilter(cmd, filter)
	}
	return printSerializedIpFilter(cmd, filter)
}

func removeIpFilterFromArray(array []string, itemToRemove string) []string {
	for i, element := range array {
		if element == itemToRemove {
			array[i] = array[len(array)-1]
			return array[:len(array)-1]
		}
	}
	return array
}
