package iam

import (
	"github.com/confluentinc/cli/v3/pkg/types"
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/log"
)

func (c *ipGroupCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an IP group.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update the name and add a CIDR block to IP group ipg-12345:`,
				Code: `confluent iam ip-group update ipg-12345 --name "New Group Name" --add-cidr-blocks 123.234.0.0/16`,
			},
		),
	}

	cmd.Flags().String("name", "", "Updated name of the IP group.")
	cmd.Flags().StringSlice("add-cidr-blocks", []string{}, "A comma-separated list of CIDR blocks to add.")
	cmd.Flags().StringSlice("remove-cidr-blocks", []string{}, "A comma-separated list of CIDR blocks to remove.")

	pcmd.AddContextFlag(cmd, c.CLICommand)
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

	currentIpGroupId := args[0]

	// Get the current IP group we are going to update
	currentIpGroup, err := c.V2Client.GetIamIpGroup(currentIpGroupId)
	if err != nil {
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

	// Initialize the IP group object that we will pass into the update command
	updateIpGroup := iamv2.IamV2IpGroup{Id: &args[0], GroupName: currentIpGroup.GroupName}
	if groupName != "" {
		updateIpGroup.GroupName = &groupName
	}

	// Create a set of the current CIDR block values
	currentCidrBlocksSet := make(types.Set[string])
	for _, cidrBlock := range currentIpGroup.GetCidrBlocks() {
		currentCidrBlocksSet.Add(cidrBlock)
	}

	// Create a set of the CIDR block values to add
	addCidrBlocksSet := make(types.Set[string])
	// Add each CIDR block to add to the set
	for _, cidrBlock := range addCidrBlocks {
		addCidrBlocksSet.Add(cidrBlock)
	}

	// Create a set of the CIDR block values to remove
	removeCidrBlocksSet := make(types.Set[string])
	for _, cidrBlock := range removeCidrBlocks {
		if addCidrBlocksSet.Contains(cidrBlock) {
			delete(addCidrBlocksSet, cidrBlock)
			log.CliLogger.Warnf("Attempting to add and remove %s.", cidrBlock)
		}
		if !currentCidrBlocksSet.Contains(cidrBlock) {
			log.CliLogger.Warnf("Attempting to remove CIDR block %s which does not exist on this IP group.", cidrBlock)
		}
		removeCidrBlocksSet.Add(cidrBlock)
	}

	// Combine the set of the current CIDR blocks and the CIDR blocks to add
	for cidrBlock := range currentCidrBlocksSet {
		// Ensure the IP group ID isn't being removed
		if !removeCidrBlocksSet.Contains(cidrBlock) {
			addCidrBlocksSet.Add(cidrBlock)
		}
	}

	// Convert the map of CIDR blocks being added into an array to make the update request
	newCidrBlocks := make([]string, 0, len(addCidrBlocksSet))
	for cidrBlock := range addCidrBlocksSet {
		newCidrBlocks = append(newCidrBlocks, cidrBlock)
	}

	updateIpGroup.CidrBlocks = &newCidrBlocks

	group, err := c.V2Client.UpdateIamIpGroup(updateIpGroup, currentIpGroupId)
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
				"Please double check the IP group you are updating."+
					" Otherwise, try again from an IP address permitted within this updated IP group or another IP group.")
		}
		return err
	}

	return printIpGroup(cmd, group)
}
