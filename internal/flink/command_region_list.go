package flink

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type regionOut struct {
	IsCurrent bool   `human:"Current" serialized:"is_current"`
	Name      string `human:"Name" serialized:"name"`
	Cloud     string `human:"Cloud" serialized:"cloud"`
	Region    string `human:"Region" serialized:"region"`
}

func (c *command) newRegionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink regions.",
		Args:  cobra.NoArgs,
		RunE:  c.regionList,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List the available Flink AWS regions.",
				Code: "confluent flink region list --cloud aws",
			},
		),
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) regionList(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	regions, err := c.V2Client.ListFlinkRegions(strings.ToUpper(cloud))
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, region := range regions {
		out := &regionOut{
			Cloud:  region.GetCloud(),
			Region: region.GetRegionName(),
			Name:   region.GetDisplayName(),
		}

		if x := strings.SplitN(region.GetId(), ".", 2); len(x) == 2 {
			out.IsCurrent = x[0] == c.Context.GetCurrentFlinkCloudProvider() && x[1] == c.Context.GetCurrentFlinkRegion()
		}

		list.Add(out)
	}
	return list.Print()
}
