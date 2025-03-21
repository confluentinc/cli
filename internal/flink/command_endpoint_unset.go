package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newEndpointUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset the current Flink endpoint.",
		Long:  "Unset the current Flink endpoint that was set with the `use` command.",
		Args:  cobra.NoArgs,
		RunE:  c.endpointUnset,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Unset the current Flink endpoint "https://flink.us-east-1.aws.confluent.cloud".`,
				Code: `confluent flink endpoint unset`,
			},
		),
	}

	return cmd
}

func (c *command) endpointUnset(_ *cobra.Command, _ []string) error {
	endpointToUnset := c.Context.GetCurrentFlinkEndpoint()
	if endpointToUnset == "" {
		return nil
	}
	if err := c.Context.SetCurrentFlinkEndpoint(""); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UnsetResourceMsg, resource.FlinkComputePool, endpointToUnset)
	return nil
}
