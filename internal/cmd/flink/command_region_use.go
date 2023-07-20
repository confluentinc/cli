package flink

import (
	"fmt"
	flinkv2 "github.com/confluentinc/ccloud-sdk-go-v2/flink/v2"
	"github.com/samber/lo"
	"github.com/spf13/cobra"
	"strings"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

func (c *command) newRegionUseCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "use <id>",
		Short:             "Use a Flink region in subsequent commands.",
		Long:              "Choose a Flink compute pool to be used in subsequent commands which support passing a compute pool with the `--compute-pool` flag.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validComputePoolArgs),
		RunE:              c.regionUse,
	}

	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) regionUse(cmd *cobra.Command, args []string) error {

	split := strings.Split(args[0], ".")
	if len(split) != 2 || split[0] == "" || split[1] == "" {
		return errors.NewErrorWithSuggestions(fmt.Sprintf("Flink region %s is invalid", args[0]), "run `ccloud flink region list` to see available regions")
	}

	regions, err := c.V2Client.ListFlinkRegions(split[0])
	if err != nil {
		return err
	}

	reg, found := lo.Find(regions, func(r flinkv2.FcpmV2Region) bool {
		return strings.ToLower(r.GetId()) == strings.ToLower(args[0])
	})
	if !found {
		return errors.NewErrorWithSuggestions(fmt.Sprintf("Flink region %s is not available", args[0]), "run `ccloud flink region list` to see available regions")
	}

	if err := c.Context.SetCurrentFlinkRegion(reg.GetRegionName()); err != nil {
		return err
	}
	if err := c.Context.SetCurrentFlinkCloudProvider(reg.GetCloud()); err != nil {
		return err
	}

	if err := c.Config.Save(); err != nil {
		return err
	}

	output.Printf(errors.UsingResourceMsg, resource.FlinkRegion, reg.GetId())
	return nil
}
