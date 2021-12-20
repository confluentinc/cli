package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <subject>",
		Short: "Update subject compatibility or mode.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.onPremUpdate),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update subject level compatibility or mode of Schema Registry:",
				Code: fmt.Sprintf("%s schema-registry subject update <subject-name> --compatibility=BACKWARD\n%s schema-registry subject update <subject-name> --mode=READWRITE %s", version.CLIName, version.CLIName, errors.OnPremAuthenticationMsg),
			},
		),
	}

	addCompatibilityFlag(cmd)
	cmd.Flags().String("sr-endpoint", "", "The URL of the schema registry cluster.")
	cmd.Flags().AddFlagSet(pcmd.OnPremKafkaRestSet())
	addModeFlag(cmd)

	return cmd
}

func (c *subjectCommand) onPremUpdate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compat != "" {
		return c.updateCompatibility(cmd, args, srClient, ctx)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateMode(cmd, args, srClient, ctx)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}
