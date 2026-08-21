package kafka

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/license"
)

func (c *licenseCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Confluent Platform license.",
		Long:  "Validate and store a Confluent Platform license.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update the license from a file.",
				Code: "confluent kafka license update --license license.jwt --url http://localhost:8090/kafka",
			},
			examples.Example{
				Text: "Validate a license without storing it.",
				Code: "confluent kafka license update --license license.jwt --dry-run --url http://localhost:8090/kafka",
			},
		),
	}

	cmd.Flags().String("license", "", "JWT-encoded license, or path to a file containing one, to validate and store.")
	cmd.Flags().Bool("dry-run", false, "Validate the license and report the resulting state without storing it.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("license"))

	return cmd
}

func (c *licenseCommand) update(cmd *cobra.Command, _ []string) error {
	restClient, restContext, clusterId, err := initKafkaRest(c.AuthenticatedCLICommand, cmd)
	if err != nil {
		return err
	}

	return license.Update(cmd, restClient, restContext, clusterId, c.Config.EnableColor)
}
