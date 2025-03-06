package flink

import (
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

	if valid := validateFlinkEndpointBeforeUse(cloud, region, endpoint); !valid {
		return errors.NewErrorWithSuggestions("endpoint doesn't match cloud provider and region selected", "Select a different endpoint, or select a cloud provider and region with `confluent flink region use` or `--cloud` and `--region` first.")
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

func validateFlinkEndpointBeforeUse(cloud, region, endpoint string) bool {
	if !strings.HasPrefix(endpoint, "https://flink") {
		return false
	}
	if cloud == "" && region == "" {
		return true
	}

	cloud = strings.ToLower(cloud)
	region = strings.ToLower(region)
	return strings.Contains(endpoint, cloud) && strings.Contains(endpoint, region)
}
