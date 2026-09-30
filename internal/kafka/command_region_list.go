package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/kafka"
	"github.com/confluentinc/cli/v4/pkg/output"
)

type regionOut struct {
	Name   string `human:"Name" serialized:"name"`
	Cloud  string `human:"Cloud" serialized:"cloud"`
	Region string `human:"Region" serialized:"region"`
}

func (c *regionCommand) newListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		Args:  cobra.NoArgs,
		RunE:  c.list,
	}

	pcmd.AddCloudFlag(cmd)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *regionCommand) list(cmd *cobra.Command, _ []string) error {
	cloud, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return err
	}

	regions, err := kafka.ListRegions(c.Client, cloud)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, region := range regions {
		out := &regionOut{
			Cloud:  region.CloudId,
			Region: region.RegionId,
			Name:   region.RegionName,
		}
		list.Add(out)
	}
	return list.Print()
}
