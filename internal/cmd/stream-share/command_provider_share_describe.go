package streamshare

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newProviderShareDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "describe <id>",
		Short:             "Describe a provider share.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validProviderShareArgs),
		RunE:              c.describeProviderShare,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe provider share "ss-12345":`,
				Code: "confluent stream-share provider share describe ss-12345",
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) describeProviderShare(cmd *cobra.Command, args []string) error {
	shareId := args[0]

	provideShare, err := c.V2Client.DescribeProviderShare(shareId)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(c.buildProviderShare(provideShare))
	return table.Print()
}
