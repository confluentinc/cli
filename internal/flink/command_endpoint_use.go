package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func (c *command) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Flink endpoint in current environment.",
		Long:  "Use a Flink endpoint in current environment for subsequent Flink dataplane commands, namely `flink connection`, `flink statement` and `flink shell`",
		Args:  cobra.ExactArgs(1),
		RunE:  c.endpointUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Use "https://flink.us-east-1.aws.confluent.cloud" for subsequent Flink dataplane commands.`,
				Code: `confluent flink endpoint use "https://flink.us-east-1.aws.confluent.cloud"`,
			},
		),
	}

	return cmd
}

func (c *command) endpointUse(_ *cobra.Command, args []string) error {
	cloud := c.Context.GetCurrentFlinkCloudProvider()
	if cloud == "" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`Current Flink cloud provider is empty`),
			"Please run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink cloud provider first.",
		)
	}

	region := c.Context.GetCurrentFlinkRegion()
	if region == "" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`Current Flink region is empty`),
			"Please run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink region first.",
		)
	}

	endpoint := args[0]
	if valid := validateUserProvidedFlinkEndpoint(endpoint, cloud, region, c); !valid {
		suggestion := fmt.Sprintf(`Please run "confluent flink endpoint list" to see all available Flink endpoints, or "confluent flink region use" to switch to a different cloud or region.`)
		return errors.NewErrorWithSuggestions(fmt.Sprintf("Flink endpoint %q is invalid for cloud = %q and region = %q", endpoint, cloud, region), suggestion)
	}

	if err := c.Context.SetCurrentFlinkEndpoint(endpoint); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkEndpoint, endpoint)
	return nil
}

// validateUserProvidedFlinkEndpoint verifies if the provided Flink endpoint is valid for the given cloud and region.
// It performs validation against three endpoint types:
// 1. Public endpoints
// 2. Private endpoints associated with PrivateLink attachments
// 3. Private endpoints associated with Confluent Cloud Networks
// Returns true if the endpoint is valid, false otherwise.
func validateUserProvidedFlinkEndpoint(endpoint, cloud, region string, c *command) bool {
	if endpoint == "" || cloud == "" || region == "" {
		log.CliLogger.Debug("Invalid input: endpoint, cloud, or region is empty")
		return false
	}

	cloud = strings.ToUpper(cloud)
	// Check if the endpoint is PUBLIC
	flinkRegions, err := c.V2Client.ListFlinkRegions(cloud, region)
	if err != nil {
		log.CliLogger.Debugf("Error listing Flink regions: %v", err)
		return false
	}

	for _, r := range flinkRegions {
		if r.GetHttpEndpoint() == endpoint {
			log.CliLogger.Debugf("Flink endpoint %q is a valid PUBLIC endpoint", endpoint)
			return true
		}
	}

	// Check if the endpoint is PRIVATE associated with PLATT
	platts, err := c.V2Client.ListPrivateLinkAttachments(c.Context.GetCurrentEnvironment(), nil, nil, nil, []string{"READY"})
	if err != nil {
		log.CliLogger.Debugf("Error listing PrivateLink attachments: %v", err)
		return false
	} else {
		filterKeyMap := buildCloudRegionKeyFilterMapFromPrivateLinkAttachments(platts)

		for _, r := range flinkRegions {
			key := CloudRegionKey{
				cloud:  r.GetCloud(),
				region: r.GetRegionName(),
			}
			if _, ok := filterKeyMap[key]; ok && r.GetPrivateHttpEndpoint() == endpoint {
				log.CliLogger.Debugf("Flink endpoint %q is a valid PRIVATE endpoint associated with a private link attachment", endpoint)
				return true
			}
		}
	}

	// Check if the endpoint is PRIVATE associated with CCN
	networks, err := c.V2Client.ListNetworks(c.Context.GetCurrentEnvironment(), nil, []string{cloud}, []string{region}, nil, []string{"READY"}, nil)
	if err != nil {
		log.CliLogger.Debugf("Error listing networks: %v", err)
		return false
	}

	for _, network := range networks {
		suffix := network.Status.GetEndpointSuffix()
		validEndpoint := fmt.Sprintf("https://flink%s", suffix)
		if endpoint == validEndpoint {
			log.CliLogger.Debugf("Flink endpoint %q is a valid PRIVATE CCN endpoint", endpoint)
			return true
		}
	}

	return false
}
