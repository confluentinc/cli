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
	"github.com/confluentinc/cli/v4/pkg/utils"
)

const resourceScopeStr = "crn://confluent.cloud/organization=%s/environment=%s"

func (c *ipFilterCommand) newCreateCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create an IP filter.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.create,
	}
	isKafkaEnabled := cfg.IsTest || (cfg.Context() != nil && featureflags.Manager.BoolVariation("auth.ip_filter.kafka.cli.enabled", cfg.Context(), featureflags.GetCcloudLaunchDarklyClient(cfg.Context().PlatformName), true, false))
	cmd.Example = examples.BuildExampleString(
		examples.Example{
			Text: `Create an IP filter named "demo-ip-filter" with operation group "management" and IP groups "ipg-12345" and "ipg-67890":`,
			Code: "confluent iam ip-filter create demo-ip-filter --operations management --ip-groups ipg-12345,ipg-67890",
		},
	)
	pcmd.AddResourceGroupFlag(cmd)
	cmd.Flags().StringSlice("ip-groups", []string{}, "A comma-separated list of IP group IDs.")
	cmd.Flags().String("environment", "", "Identifier of the environment for which this filter applies. Without this flag, applies only to the organization.")
	opGroups := []string{"MANAGEMENT", "SCHEMA", "FLINK"}
	if isKafkaEnabled {
		opGroups = append(opGroups, "KAFKA_MANAGEMENT", "KAFKA_DATA")
	}
	cmd.Flags().StringSlice("operations", nil, fmt.Sprintf("A comma-separated list of operation groups: %s.", utils.ArrayToCommaDelimitedString(opGroups, "or")))
	cmd.Flags().Bool("no-public-networks", false, "Use in place of ip-groups to reference the no public networks IP Group.")
	cmd.MarkFlagsMutuallyExclusive("ip-groups", "no-public-networks")
	cmd.MarkFlagsMutuallyExclusive("resource-group", "operations")
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

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
	resourceScope := ""
	operationGroups := []string{}
	orgId := c.Context.GetCurrentOrganization()
	environment, err := cmd.Flags().GetString("environment")
	if err != nil {
		return err
	}
	if environment != "" {
		resourceScope = fmt.Sprintf(resourceScopeStr, orgId, environment)
	}
	operationGroups, err = cmd.Flags().GetStringSlice("operations")
	if err != nil {
		return err
	}
	if len(operationGroups) == 0 && resourceGroup == "multiple" {
		operationGroups = []string{"MANAGEMENT"}
	}
	if resourceGroup == "management" {
		operationGroups = []string{"MANAGEMENT"}
		resourceGroup = "multiple"
	}
	npnGroup, err := cmd.Flags().GetBool("no-public-networks")
	if err != nil {
		return err
	}
	if npnGroup {
		ipGroups = []string{"ipg-none"}
	}

	// Convert the IP group IDs into IP group objects
	ipGroupIdObjects := make([]iamipfilteringv2.GlobalObjectReference, len(ipGroups))
	for i, ipGroupId := range ipGroups {
		// The empty string fields will get filled in automatically by the cc-policy-service
		ipGroupIdObjects[i] = iamipfilteringv2.GlobalObjectReference{Id: ipGroupId}
	}

	createIpFilter := iamipfilteringv2.IamV2IpFilter{
		FilterName:      &args[0],
		ResourceGroup:   &resourceGroup,
		IpGroups:        &ipGroupIdObjects,
		ResourceScope:   &resourceScope,
		OperationGroups: &operationGroups,
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
