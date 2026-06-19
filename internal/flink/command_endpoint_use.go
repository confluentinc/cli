package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Flink endpoint.",
		Long:  "Use a Flink endpoint as active endpoint for all subsequent Flink dataplane commands in current environment, such as `flink connection`, `flink statement` and `flink shell`.",
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
			"Current Flink cloud provider is empty",
			"Please run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink cloud provider first.",
		)
	}

	region := c.Context.GetCurrentFlinkRegion()
	if region == "" {
		return errors.NewErrorWithSuggestions(
			"Current Flink region is empty",
			"Please run `confluent flink region use --cloud <cloud> --region <region>` to set the Flink region first.",
		)
	}

	endpoint := args[0]
	valid, err := validateUserProvidedFlinkEndpoint(endpoint, cloud, region, c)
	if err != nil {
		return err
	}
	if !valid {
		suggestion := `Please run "confluent flink endpoint list" to see all available Flink endpoints, or "confluent flink region use" to switch to a different cloud or region.`
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

// validateUserProvidedFlinkEndpoint checks whether the given endpoint URL is one of
// the REST endpoints reported by the Endpoints API for the current environment +
// cloud + region. Returns (true, nil) on a match, (false, nil) if the endpoint is
// not in the returned set, or (false, err) if the environment lookup or API call
// fails — so the caller can surface a real error instead of a misleading
// "endpoint invalid" message.
func validateUserProvidedFlinkEndpoint(endpoint, cloud, region string, c *command) (bool, error) {
	if endpoint == "" {
		return false, nil
	}

	environmentId, err := c.Context.EnvironmentId()
	if err != nil {
		return false, err
	}

	endpoints, err := c.V2Client.ListEndpoints(environmentId, cloud, region, flinkEndpointService, nil, "")
	if err != nil {
		return false, err
	}

	for _, e := range endpoints {
		// Skip LANGUAGE_SERVICE endpoints (`flinkpls.*`) — they are used by the Cloud
		// Console SQL editor's language server, not by CLI dataplane commands.
		if e.GetEndpointType() != flinkRestEndpointType {
			continue
		}
		if flinkEndpointUrl(e.GetEndpoint()) == endpoint {
			return true, nil
		}
	}
	return false, nil
}
