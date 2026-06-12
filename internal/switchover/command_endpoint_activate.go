package switchover

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
)

func (c *command) newEndpointActivateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:               "activate <id>",
		Short:             "Activate a switchover endpoint.",
		Long:              "Activate a switchover endpoint, applying its desired routing target.",
		Args:              cobra.ExactArgs(1),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validEndpointArgs),
		RunE:              c.endpointActivate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Activate switchover endpoint "se-123456".`,
				Code: "confluent switchover endpoint activate se-123456",
			},
		),
	}

	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) endpointActivate(cmd *cobra.Command, args []string) error {
	endpoint, err := c.V2Client.ActivateSwitchoverEndpoint(args[0])
	if err != nil {
		return err
	}

	return printEndpointTable(cmd, endpoint)
}
