package endpoint

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "activate <id>",
		Short: "Activate a switchover endpoint.",
		Long:  "Activate a switchover endpoint, applying its desired routing target.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.activate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Activate switchover endpoint "se-123456".`,
				Code: `confluent switchover endpoint activate se-123456`,
			},
		),
	}

	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) activate(cmd *cobra.Command, args []string) error {
	result, err := c.V2Client.ActivateSwitchoverEndpoint(args[0])
	if err != nil {
		return err
	}

	return printSwitchoverEndpoint(cmd, result)
}
