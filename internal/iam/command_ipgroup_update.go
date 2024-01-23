package iam

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/types"
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
				Text: "Update the name and add a CIDR block to IP group `ipg-12345`:",
				Code: `confluent iam ip-group update ipg-12345 --name "New Group Name" --add-cidr-blocks 123.234.0.0/16`,
			},
		),
	}

	cmd.Flags().String("name", "", "Updated name of the IP group.")
	cmd.Flags().StringSlice("add-cidr-blocks", []string{}, "A comma-separated list of CIDR blocks to add.")
	cmd.Flags().StringSlice("remove-cidr-blocks", []string{}, "A comma-separated list of CIDR blocks to remove.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cmd.MarkFlagsOneRequired("name", "add-cidr-blocks", "remove-cidr-blocks")

	return cmd
}

func (c *ipGroupCommand) update(cmd *cobra.Command, args []string) error {
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
	updateIpGroup := iamv2.IamV2IpGroup{
		Id:        &args[0],
		GroupName: currentIpGroup.GroupName,
	}

	if groupName != "" {
		updateIpGroup.GroupName = &groupName
	}

	newCidrBlocks, warnings := types.AddAndRemove(currentIpGroup.GetCidrBlocks(), addCidrBlocks, removeCidrBlocks)
	for _, warning := range warnings {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] %s\n", warning)
	}

	if len(newCidrBlocks) == 0 {
		return errors.NewErrorWithSuggestions("Cannot remove all CIDR blocks from IP group",
			fmt.Sprintf("Please double check the IP group you are updating. Use `confluent iam ip-group describe %s` to see the CIDR blocks associated with an IP group.", currentIpGroup.GetId()))
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
			return errors.NewErrorWithSuggestions(
				err.Error()[:errorMessageIndex-1],
				"Please double check the IP group you are updating. Otherwise, try again from an IP address permitted within this updated IP group or another IP group.",
			)
		}
		return err
	}

	return printIpGroup(cmd, group)
}
