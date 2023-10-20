package flink

import (
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

type regionOut struct {
	Id         string `human:"ID" serialized:"id"`
	Name       string `human:"Name" serialized:"name"`
	Cloud      string `human:"Cloud" serialized:"cloud"`
	RegionName string `human:"Region Name" serialized:"region_name"`
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
		list.Add(&regionOut{
			Id:         region.GetId(),
			Name:       region.GetDisplayName(),
			Cloud:      region.GetCloud(),
			RegionName: region.GetRegionName(),
		})
	}
	return list.Print()
}
