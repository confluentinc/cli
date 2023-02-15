package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newProviderShareListCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List shares for provider.",
		Args:  cobra.NoArgs,
		RunE:  c.listProviderShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `List provider shares for shared resource "sr-12345":`,
				Code: "confluent stream-share provider share list --shared-resource sr-12345",
			},
		),
	}

	cmd.Flags().String("shared-resource", "", "Filter the results by exact match for shared resource.")

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) listProviderShare(cmd *cobra.Command, _ []string) error {
	sharedResource, err := cmd.Flags().GetString("shared-resource")
	if err != nil {
		return err
	}

	providerShares, err := c.V2Client.ListProviderShares(sharedResource)
	if err != nil {
		return err
	}

	list := output.NewList(cmd)
	for _, share := range providerShares {
		list.Add(c.buildProviderShare(share))
	}
	return list.Print()
}
