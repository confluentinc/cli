package price

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/kafka"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	poutput "github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) autocompleteFlags(cmd *cobra.Command) {
	pcmd.RegisterFlagCompletionFunc(cmd, "cloud", func(_ *cobra.Command, _ []string) []string { return clouds })
	pcmd.RegisterFlagCompletionFunc(cmd, "availability", func(_ *cobra.Command, _ []string) []string { return availabilities })
	pcmd.RegisterFlagCompletionFunc(cmd, "cluster-type", func(_ *cobra.Command, _ []string) []string { return clusterTypes })
	pcmd.RegisterFlagCompletionFunc(cmd, "network-type", func(_ *cobra.Command, _ []string) []string { return networkTypes })
	pcmd.RegisterFlagCompletionFunc(cmd, "metric", func(_ *cobra.Command, _ []string) []string { return metrics })
	pcmd.RegisterFlagCompletionFunc(cmd, poutput.FlagName, func(_ *cobra.Command, _ []string) []string { return poutput.ValidFlagValues })

	pcmd.RegisterFlagCompletionFunc(cmd, "region", func(cmd *cobra.Command, args []string) []string {
		if err := c.PersistentPreRunE(cmd, args); err != nil {
			return nil
		}

		cloud, _ := cmd.Flags().GetString("cloud")
		return c.autocompleteRegions(cloud)
	})
}

func (c *command) autocompleteRegions(cloud string) []string {
	regions, err := kafka.ListRegions(c.Client, cloud)
	if err != nil {
		return nil
	}

	suggestions := make([]string, len(regions))
	for i, region := range regions {
		suggestions[i] = region.RegionId
	}
	return suggestions
}
