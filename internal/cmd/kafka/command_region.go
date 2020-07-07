package kafka

import (
	"context"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
)

var (
	regionListFields           = []string{"CloudId", "CloudName", "RegionId", "RegionName"}
	regionListHumanLabels      = []string{"CloudId", "CloudName", "RegionId", "RegionName"}
	regionListStructuredLabels = []string{"cloud_id", "cloud_name", "region_id", "region_name"}
)

type regionCommand struct {
	*pcmd.AuthenticatedCLICommand
}

// NewRegionCommand returns the Cobra command for Kafka region.
func NewRegionCommand(prerunner pcmd.PreRunner) *cobra.Command {
	cliCmd := pcmd.NewAuthenticatedCLICommand(
		&cobra.Command{
			Use:   "region",
			Short: "Manage Confluent Cloud regions.",
			Long:  "Use this command to manage Confluent Cloud regions.",
		}, prerunner)
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
	listCmd.Flags().StringP(output.FlagName, output.ShortHandFlag, output.DefaultValue, output.Usage)
	listCmd.Flags().SortFlags = false
	c.AddCommand(listCmd)
}

func (c *regionCommand) list(cmd *cobra.Command, _ []string) error {
	clouds, err := c.Client.EnvironmentMetadata.Get(context.Background())
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	cloudIdFilter, err := cmd.Flags().GetString("cloud")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	outputWriter, err := output.NewListOutputWriter(cmd, regionListFields, regionListHumanLabels, regionListStructuredLabels)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}
	type regionStruct struct {
		CloudId    string
		CloudName  string
		RegionId   string
		RegionName string
	}
	for _, cloud := range clouds {
		for _, region := range cloud.Regions {
			if !region.IsSchedulable || (cloudIdFilter != "" && cloudIdFilter != cloud.Id) {
				continue
			}
			outputWriter.AddElement(&regionStruct{
				CloudId:    cloud.Id,
				CloudName:  cloud.Name,
				RegionId:   region.Id,
				RegionName: region.Name,
			})
		}
	}
	return outputWriter.Out()
}
