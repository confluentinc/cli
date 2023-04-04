package schemaregistry

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
)

func (c *command) newSubjectUpdateCommandOnPrem() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update <subject>",
		Short:       "Update subject compatibility or mode.",
		Args:        cobra.ExactArgs(1),
		RunE:        c.subjectUpdateOnPrem,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireOnPremLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update subject-level compatibility of subject "payments".`,
				Code: fmt.Sprintf("confluent schema-registry subject update payments --compatibility backward %s", OnPremAuthenticationMsg),
			},
			examples.Example{
				Text: `Update subject-level mode of subject "payments".`,
				Code: fmt.Sprintf("confluent schema-registry subject update payments --mode readwrite %s", OnPremAuthenticationMsg),
			},
		),
	}

	addCompatibilityFlag(cmd)
	cmd.Flags().String("compatibility-group", "", "The name for compatibility group.")
	cmd.Flags().String("default-metadata", "", "The path to default metadata file.")
	cmd.Flags().String("override-metadata", "", "The path to override metadata file.")
	cmd.Flags().String("default-ruleset", "", "The path to default schema ruleset file.")
	cmd.Flags().String("override-ruleset", "", "The path to override schema ruleset file.")
	addModeFlag(cmd)
	cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	pcmd.AddContextFlag(cmd, c.CLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("default-metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("override-metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("default-ruleset", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("override-ruleset", "json"))

	return cmd
}

func (c *command) subjectUpdateOnPrem(cmd *cobra.Command, args []string) error {
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
		return c.updateCompatibility(subject, compatibility, cmd, srClient, ctx)
	}

	if mode != "" {
		return c.updateMode(subject, mode, srClient, ctx)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}
