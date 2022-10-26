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
		Use:         "update <subject>",
		Short:       "Update subject compatibility or mode.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.onPremUpdate,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update subject level compatibility of subject "payments"`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --compatibility=BACKWARD %s", version.CLIName, OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: `Update subject level mode of subject "payments"`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --mode=READWRITE %s", version.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	addCompatibilityFlag(cmd)
	addModeFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *subjectCommand) onPremUpdate(cmd *cobra.Command, args []string) error {
	subject := args[0]

	srClient, ctx, err := GetSrApiClientWithToken(cmd, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	if compatibility != "" && mode != "" {
		return errors.New(errors.CompatibilityOrModeErrorMsg)
	}

	if compatibility != "" {
		return c.updateCompatibility(cmd, subject, compatibility, srClient, ctx)
	}

	if mode != "" {
		return c.updateMode(cmd, subject, mode, srClient, ctx)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}
