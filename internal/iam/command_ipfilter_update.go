package iam

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/config"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/featureflags"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/types"
)

func (c *ipFilterCommand) newUpdateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "update <id>",
		Short:             "Update an IP filter.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
	}
	isKafkaEnabled := cfg.IsTest || (cfg.Context() != nil && featureflags.Manager.BoolVariation("auth.ip_filter.kafka.cli.enabled", cfg.Context(), featureflags.GetCcloudLaunchDarklyClient(cfg.Context().PlatformName), true, false))
	operationGroups := []string{"SCHEMA", "FLINK"}
	if isKafkaEnabled {
		operationGroups = append(operationGroups, "KAFKA_MANAGEMENT", "KAFKA_DATA")
	}
	cmd.Example = examples.BuildExampleString(
		examples.Example{
			Text: `Update the name and add an IP group and operation group to IP filter "ipf-abcde":`,
			Code: fmt.Sprintf(`confluent iam ip-filter update ipf-abcde --name "New Filter Name" --add-ip-groups ipg-12345 --add-operation-groups %s`, strings.Join(operationGroups, ",")),
		},
	)
	cmd.Flags().String("name", "", "Updated name of the IP filter.")
	pcmd.AddResourceGroupFlag(cmd)

	cmd.Flags().StringSlice("add-ip-groups", []string{}, "A comma-separated list of IP groups to add.")
	cmd.Flags().StringSlice("remove-ip-groups", []string{}, "A comma-separated list of IP groups to remove.")
	cmd.Flags().StringSlice("add-operation-groups", []string{}, "A comma-separated list of operation groups to add.")
	cmd.Flags().StringSlice("remove-operation-groups", []string{}, "A comma-separated list of operation groups to remove.")
	cmd.MarkFlagsOneRequired("name", "resource-group", "add-ip-groups", "remove-ip-groups", "add-operation-groups", "remove-operation-groups")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *ipFilterCommand) update(cmd *cobra.Command, args []string) error {
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

	currentOperationGroups := currentIpFilter.GetOperationGroups()

	// Initialize our update IP filter object with the current IP filter values
	updateIpFilter := currentIpFilter

	if filterName != "" {
		updateIpFilter.FilterName = &filterName
	}

	if resourceGroup != "" {
		updateIpFilter.ResourceGroup = &resourceGroup
	}

	newIpGroupIds, warnings := types.AddAndRemove(currentIpGroupIds, addIpGroups, removeIpGroups)
	for _, warning := range warnings {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] %s\n", warning)
	}

	// Convert the IP group IDs into IP group objects
	IpGroupIdObjects := make([]iamipfilteringv2.GlobalObjectReference, len(newIpGroupIds))
	for i, ipGroupId := range newIpGroupIds {
		IpGroupIdObjects[i] = iamipfilteringv2.GlobalObjectReference{Id: ipGroupId}
	}

	if len(IpGroupIdObjects) == 0 {
		return errors.NewErrorWithSuggestions("Cannot remove all IP groups from IP filter",
			fmt.Sprintf("Please double check the IP filter you are updating. Use `confluent iam ip-filter describe %s` to see the IP groups associated with an IP filter.", currentIpFilter.GetId()))
	}

	updateIpFilter.IpGroups = &IpGroupIdObjects
	addOperationGroups, err := cmd.Flags().GetStringSlice("add-operation-groups")
	if err != nil {
		return err
	}

	removeOperationGroups, err := cmd.Flags().GetStringSlice("remove-operation-groups")
	if err != nil {
		return err
	}
	newOperationGroups, warnings := types.AddAndRemove(currentOperationGroups, addOperationGroups, removeOperationGroups)
	for _, warning := range warnings {
		output.ErrPrintf(c.Config.EnableColor, "[WARN] %s\n", warning)
	}
	if len(newOperationGroups) == 0 && resourceGroup == "multiple" {
		newOperationGroups = []string{"MANAGEMENT"}
	}
	updateIpFilter.SetOperationGroups(newOperationGroups)

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
			return errors.NewErrorWithSuggestions(
				err.Error()[:errorMessageIndex-1],
				"Please double check the IP filter you are updating. Otherwise, try again from an IP address permitted within this updated IP filter or another IP filter.",
			)
		}
		return err
	}

	return printIpFilter(cmd, filter)
}
