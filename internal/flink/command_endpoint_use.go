package flink

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
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
	region := c.Context.GetCurrentFlinkRegion()
	endpoint := args[0]

	if valid := validateFlinkEndpointFormat(endpoint); !valid {
		suggestion := fmt.Sprintf(`Please use "confluent flink endpoint list" to select a valid Flink endpoint starting with "https://flink"`)
		return errors.NewErrorWithSuggestions("this endpoint format is invalid", suggestion)
	}

	if valid := validateFlinkEndpointMatchCloudAndRegion(cloud, region, endpoint); !valid {
		suggestion := fmt.Sprintf(`Please use "confluent flink endpoint list" to select a different endpoint to match cloud = %s and region = %s, or select a different cloud provider and region with "confluent flink region use" command to match this endpoint`, cloud, region)
		return errors.NewErrorWithSuggestions("this endpoint doesn't match cloud provider and region selected", suggestion)
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

func validateFlinkEndpointFormat(endpoint string) bool {
	if !strings.HasPrefix(endpoint, "https://flink") {
		return false
	}
	return true
}

func validateFlinkEndpointMatchCloudAndRegion(cloud, region, endpoint string) bool {
	if cloud == "" && region == "" {
		return true
	}

	cloud = strings.ToLower(cloud)
	region = strings.ToLower(region)
	return strings.Contains(endpoint, cloud) && strings.Contains(endpoint, region)
}
