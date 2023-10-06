package flink

import (
	"fmt"
	"strings"

	"github.com/samber/lo"
	"github.com/spf13/cobra"

	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
)

func (c *command) newRegionUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Use a Flink region in subsequent commands.",
		Long:              "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validRegionArgs),
		RunE:              c.regionUse,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *command) validRegionArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	regions, err := c.V2Client.ListFlinkRegions("")
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = fmt.Sprintf("%s\t%s", region.GetId(), region.GetDisplayName())
	}
	return suggestions
}

func (c *command) regionUse(cmd *cobra.Command, args []string) error {
	split := strings.Split(args[0], ".")
	if len(split) != 2 || split[0] == "" || split[1] == "" {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`Flink region "%s" is invalid`, args[0]),
			"Run `confluent flink region list` to see available regions.",
		)
	}

	regions, err := c.V2Client.ListFlinkRegions(split[0])
	if err != nil {
		return err
	}

	region, ok := lo.Find(regions, func(region flinkv2.FcpmV2Region) bool {
		return strings.EqualFold(region.GetId(), args[0])
	})
	if !ok {
		return errors.NewErrorWithSuggestions(
			fmt.Sprintf(`Flink region "%s" is not available`, args[0]),
			"Run `confluent flink region list` to see available regions.",
		)
	}

	if err := c.Context.SetCurrentFlinkRegion(region.GetRegionName()); err != nil {
		return err
	}

	if err := c.Context.SetCurrentFlinkCloudProvider(region.GetCloud()); err != nil {
		return err
	}

	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, errors.UsingResourceMsg, resource.FlinkRegion, region.GetId())
	return nil
}
