package kafka

import (
	"context"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	regionListLabels = []string{"CloudId", "CloudName", "RegiondId", "RegionName"}
)


type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewRegionCommand returns the Cobra command for Kafka region.
func NewRegionCommand(prerunner pcmd.PreRunner, config *v2.Config) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "region",
			Short: "Cloud regions.",
		},
		config, prerunner)
	cmd := &regionCommand{
		AuthenticatedCLICommand: cliCmd,
	}
	cmd.init()
	return cmd.Command
}

func (c *regionCommand) init() {
	listCmd := &cobra.Command{
		Use:   "list",
		Short: "List cloud provider regions.",
		RunE:  c.list,
		Args:  cobra.NoArgs,
	}
	listCmd.Flags().String("cloud", "", "The cloud ID to filter by.")
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)
}

func (c *regionCommand) list(cmd *cobra.Command, args []string) error {
	clouds, err := c.Client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cloudIdFilter, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	var data [][]string
	for _, cloud := range clouds {
		for _, region := range cloud.Regions {
			if !region.IsSchedulable || (cloudIdFilter != "" && cloudIdFilter != cloud.Id) {
				continue
			}
			row := []string{cloud.Id, cloud.Name}
			row = append(row, region.Id, region.Name)
			data = append(data, row)
		}
	}
	printer.RenderCollectionTable(data, regionListLabels)
	return nil
}
