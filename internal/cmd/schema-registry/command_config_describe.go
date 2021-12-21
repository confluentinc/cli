package schemaregistry

import (
	"fmt"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	pversion "github.com/confluentinc/cli/internal/pkg/version"
	"github.com/spf13/cobra"
)

func (c *configCommand) newDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe <subject>",
		Short: "Describe the config of a subject, or at global level.",
		Args:  cobra.MaximumNArgs(1),
		RunE:  pcmd.NewCLIRunE(c.describe),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Describe the config of a given-name subject.",
				Code: fmt.Sprintf("%s schema-registry config describe <subject-name> %s", pversion.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().StringP("subject", "S", "", SubjectUsage)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *configCommand) describe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}
