package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <subject>",
		Short:       "Update subject compatibility or mode.",
		Args:        cobra.ExactArgs(1),
		RunE:        pcmd.NewCLIRunE(c.onPremUpdate),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update subject level compatibility or mode of subject "payments"`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --compatibility=BACKWARD %s\n%s schema-registry subject update payments --mode=READWRITE %s", version.CLIName, OnPremAuthenticationMsg, version.CLIName, OnPremAuthenticationMsg),
			},
		),
	}

	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	addCompatibilityFlag(cmd)
	addModeFlag(cmd)

	return cmd
}

func (c *subjectCommand) onPremUpdate(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetAPIClientWithToken(cmd, nil, c.Version, c.AuthToken())
	if err != nil {
		return err
	}

	return c.updateSchemaSubject(cmd, args, srClient, ctx)
}
