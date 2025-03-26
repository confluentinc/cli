package flink

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
)

func (c *command) newRegionUnsetCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Unset the current Flink region.",
		Long:  "Unset the current Flink region that was set with the `use` command.",
		Args:  cobra.NoArgs,
		RunE:  c.regionUnset,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Unset the current Flink region us-west-1 with cloud provider = AWS.`,
				Code: `confluent flink region unset`,
			},
		),
	}

	return cmd
}

func (c *command) regionUnset(_ *cobra.Command, _ []string) error {
	regionToUnset := c.Context.GetCurrentFlinkRegion()
	cloudToUnset := c.Context.GetCurrentFlinkCloudProvider()
	output.Println(c.Config.EnableColor, "Unset the current Flink region")

	if cloudToUnset == "" && regionToUnset == "" {
		return nil
	}
	if err := c.Context.SetCurrentFlinkCloudProvider(""); err != nil {
		return err
	}
	if err := c.Context.SetCurrentFlinkRegion(""); err != nil {
		return err
	}
	if err := c.Config.Save(); err != nil {
		return err
	}

	return nil
}
