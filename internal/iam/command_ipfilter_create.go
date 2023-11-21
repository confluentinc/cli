package iam

import (
	"strings"

	"github.com/spf13/cobra"

	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
)

func (c *ipFilterCommand) newCreateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an IP filter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Create an IP filter named "demo-ip-filter" with resource group "management" and IP groups "ipg-12345" and "ipg-67890":`,
				Code: "confluent iam ip-filter create demo-ip-filter --resource-group management --ip-groups ipg-12345,ipg-67890",
			},
		),
	}

	pcmd.AddResourceGroupFlag(cmd)
	cmd.Flags().StringSlice("ip-groups", []string{}, "A comma-separated list of IP group IDs.")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("resource-group"))
	cobra.CheckErr(cmd.MarkFlagRequired("ip-groups"))

	return cmd
}

func (c *ipFilterCommand) create(cmd *cobra.Command, args []string) error {
	resourceGroup, err := cmd.Flags().GetString("resource-group")
	if err != nil {
		return err
	}

	ipGroups, err := cmd.Flags().GetStringSlice("ip-groups")
	if err != nil {
		return err
	}

	// Convert the IP group IDs into IP group objects
	ipGroupIdObjects := make([]iamv2.GlobalObjectReference, len(ipGroups))
	for i, ipGroupId := range ipGroups {
		// The empty string fields will get filled in automatically by the cc-policy-service
		ipGroupIdObjects[i] = iamv2.GlobalObjectReference{Id: ipGroupId}
	}

	createIpFilter := iamv2.IamV2IpFilter{
		FilterName:    &args[0],
		ResourceGroup: &resourceGroup,
		IpGroups:      &ipGroupIdObjects,
	}

	filter, err := c.V2Client.CreateIamIpFilter(createIpFilter)
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
				"Please double check the IP filter you are creating. Otherwise, try again from an IP address permitted within this IP filter.",
			)
		}
		return err
	}

	return printIpFilter(cmd, filter)
}
