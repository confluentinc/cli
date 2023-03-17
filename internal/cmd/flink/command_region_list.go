package flink

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type regionOut struct {
	Id    string `human:"ID" serialized:"id"`
	Name  string `human:"Name" serialized:"name"`
	Cloud string `human:"Cloud" serialized:"cloud"`
}

func (c *command) newRegionListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Flink regions.",
		Args:  cobra.NoArgs,
		RunE:  c.regionList,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) regionList(cmd *cobra.Command, args []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	regions, err := c.V2Client.ListFlinkRegions(cloud)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, region := range regions {
		list.Add(&regionOut{
			Id:    region.GetId(),
			Name:  region.GetDisplayName(),
			Cloud: region.GetCloud(),
		})
	}
	return list.Print()
}
