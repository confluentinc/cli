package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/resource"
)

func (c *command) newEndpointUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Flink endpoint in subsequent commands.",
		Long:  "Choose a Flink endpoint to be used in subsequent commands which support passing a region with the `--region` flag.",
		Args:  cobra.NoArgs,
		RunE:  c.endpointUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Select region "xxx" for use in subsequent Flink commands.`,
				Code: "confluent flink endpoint use xxx",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddRegionFlagFlink(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) endpointUse(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	region, err := cmd.Flags().GetString("region")
	if err != nil {
		return err
	}

	regions, err := c.V2Client.ListFlinkRegions(cloud)
	if err != nil {
		return err
	}

	var currentRegion *flinkv2.FcpmV2Region
	for _, r := range regions {
		r := r
		if r.GetRegionName() == region {
			currentRegion = &r
			break
		}
	}
	if currentRegion == nil {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`Flink region "%s" is not available for cloud provider "%s"`, region, cloud),
			"Run `confluent flink region list` to see available regions.",
		)
	}

	if err := c.Context.SetCurrentFlinkCloudProvider(cloud); err != nil {
		return err
	}

	if err := c.Context.SetCurrentFlinkRegion(region); err != nil {
		return err
	}

	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkRegion, currentRegion.GetDisplayName())
	return nil
}
