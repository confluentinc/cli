package flink

import (
	"fmt"

	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newRegionUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "use",
		Short: "Use a Flink region in subsequent commands.",
		Long:  "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:  cobra.NoArgs,
		RunE:  c.regionUse,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Select the N. Virginia (us-east-1) region for use in subsequent Flink commands.",
				Code: "confluent flink region use --cloud aws --region us-east-1",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	c.addRegionFlag(cmd)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagRequired("cloud"))
	cobra.CheckErr(cmd.MarkFlagRequired("region"))

	return cmd
}

func (c *command) regionUse(cmd *cobra.Command, _ []string) error {
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
