package iam

import (
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/log"
	"github.com/confluentinc/cli/v3/pkg/types"
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
				Text: `Update the name and add an IP group to IP filter ipf-abcde:`,
				Code: `confluent iam ip-filter update ipf-abcde --name "New Filter Name" --add-ip-groups ipg-12345`,
			},
		),
	}

	cmd.Flags().String("name", "", "Updated name of the IP filter.")
	pcmd.AddResourceGroupFlag(cmd)
	cmd.Flags().StringSlice("add-ip-groups", []string{}, "A comma-separated list of IP groups to add.")
	cmd.Flags().StringSlice("remove-ip-groups", []string{}, "A comma-separated list of IP groups to remove.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
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

	currentIpFilterId := args[0]

	// Get the current IP filter we are going to update
	currentIpFilter, err := c.V2Client.GetIamIpFilter(currentIpFilterId)
	if err != nil {
		return err
	}

	// Initialize our new IP groups list with the existing ids
	currentIpGroupIds := convertIpGroupsToIds(currentIpFilter.GetIpGroups())

	// Initialize our update IP filter object with the current IP filter values
	updateIpFilter := currentIpFilter

	if filterName != "" {
		updateIpFilter.FilterName = &filterName
	}

	if resourceGroup != "" {
		updateIpFilter.ResourceGroup = &resourceGroup
	}

	// Create a set of the current IP group IDs
	currentIpGroupIdsSet := make(types.Set[string])
	for _, ipGroupId := range currentIpGroupIds {
		currentIpGroupIdsSet.Add(ipGroupId)
	}

	// Create a set of the new IP group ID values to add
	addIpGroupIdsSet := make(types.Set[string])
	// Add each new IP Group ID to the set
	for _, ipGroupId := range addIpGroups {
		if currentIpGroupIdsSet.Contains(ipGroupId) {
			AddDuplicateResource(ipGroupId, log.CliLogger)
		}
		addIpGroupIdsSet.Add(ipGroupId)
	}

	// Create a set of the new IP group ID values to remove
	removeIpGroupIdsSet := make(types.Set[string])
	for _, ipGroupId := range removeIpGroups {
		if addIpGroupIdsSet[ipGroupId] {
			delete(addIpGroupIdsSet, ipGroupId)
			AddAndDeleteResource(ipGroupId, log.CliLogger)
		}
		if !currentIpGroupIdsSet.Contains(ipGroupId) {
			DeleteNonExistentResource(ipGroupId, log.CliLogger)
		}
		removeIpGroupIdsSet[ipGroupId] = true
	}

	// Combine the map of the current IP group IDs and the IP group IDs to add
	for ipGroupId := range currentIpGroupIdsSet {
		// Ensure the IP group ID isn't being removed
		if !removeIpGroupIdsSet.Contains(ipGroupId) {
			addIpGroupIdsSet.Add(ipGroupId)
		}
	}

	// Convert the map of IP group IDs being added into an array to make the update request
	newIpGroupIds := make([]string, 0, len(addIpGroupIdsSet))
	for ipGroupId := range addIpGroupIdsSet {
		newIpGroupIds = append(newIpGroupIds, ipGroupId)
	}

	// Convert the IP group IDs into IP group objects
	IpGroupIdObjects := make([]iamv2.GlobalObjectReference, len(newIpGroupIds))
	for i, ipGroupId := range newIpGroupIds {
		IpGroupIdObjects[i] = iamv2.GlobalObjectReference{Id: ipGroupId}
	}

	updateIpFilter.IpGroups = &IpGroupIdObjects

	filter, err := c.V2Client.UpdateIamIpFilter(updateIpFilter, currentIpFilterId)
	if err != nil {
		/*
		 * Unique error message for deleting an IP Filter that would lock out the user.
		 * Splits the error message into its two components of the error and the suggestion.
		 *
		 * This uses err.Error() rather than creating its own string, because the user's
		 * IP information is inside of err.Error() string
		 *
		 * err.Error() would look like:
		 * "this action would lock out the requester from IP address <ip-address>. Please ..."
		 */
		if strings.Contains(err.Error(), "lock out") {
			errorMessageIndex := strings.Index(err.Error(), "Please")
			return errors.NewErrorWithSuggestions(err.Error()[:errorMessageIndex-1],
				"Please double check the IP filter you are updating."+
					" Otherwise, try again from an IP address permitted within this updated IP filter or another IP filter.")
		}
		return err
	}

	return printIpFilter(cmd, filter)
}
