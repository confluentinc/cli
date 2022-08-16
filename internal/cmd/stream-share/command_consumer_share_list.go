package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newListConsumerSharesCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shares for consumer.",
		Args:  cobra.NoArgs,
		RunE:  c.listConsumerShares,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "List consumer shares:",
				Code: "confluent stream-share consumer share list",
			},
		),
	}

	cmd.Flags().String("shared-resource", "", "Filter the results by exact match for shared resource.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listConsumerShares(cmd *cobra.Command, _ []string) error {
	sharedResource, err := cmd.Flags().GetString("shared-resource")
	if err != nil {
		return err
	}

	consumerShares, err := c.V2Client.ListConsumerShares(sharedResource)
	if err != nil {
		return err
	}

	outputWriter, err := output.NewListOutputWriter(cmd, consumerShareListFields, consumerShareListHumanLabels, consumerShareListStructuredLabels)
	if err != nil {
		return err
	}

	for _, share := range consumerShares {
		element := c.buildConsumerShare(share)
		outputWriter.AddElement(element)
	}

	return outputWriter.Out()
}
