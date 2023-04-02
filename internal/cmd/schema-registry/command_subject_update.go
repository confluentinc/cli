package schemaregistry

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *command) newSubjectUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <subject>",
		Short: "Update subject compatibility or mode.",
		Args:  cobra.ExactArgs(1),
		RunE:  c.subjectUpdate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update subject-level compatibility of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --compatibility backward", version.CLIName),
			},
			examples.Example{
				Text: `Update subject-level mode of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --mode readwrite", version.CLIName),
			},
		),
	}

	addCompatibilityFlag(cmd)
	cmd.Flags().String("compatibility-group", "", "The name for compatibility group.")
	cmd.Flags().String("default-metadata", "", "The path to default metadata file.")
	cmd.Flags().String("override-metadata", "", "The path to override metadata file.")
	cmd.Flags().String("default-rule-set", "", "The path to default schema rule set file.")
	cmd.Flags().String("override-rule-set", "", "The path to override schema rule set file.")
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	cobra.CheckErr(cmd.MarkFlagFilename("default-metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("override-metadata", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("default-rule-set", "json"))
	cobra.CheckErr(cmd.MarkFlagFilename("override-rule-set", "json"))

	return cmd
}

func (c *command) subjectUpdate(cmd *cobra.Command, args []string) error {
	subject := args[0]

	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
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

func (c *command) updateCompatibility(subject, compatibility string, cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {

	compatibilityGroup, err := cmd.Flags().GetString("compatibility-group")
	if err != nil {
		return err
	}

	defaultMetadata, err := ReadMetadata("default-metadata", cmd)
	if err != nil {
		return err
	}

	overrideMetadata, err := ReadMetadata("override-metadata", cmd)
	if err != nil {
		return err
	}

	defaultRuleSet, err := ReadRuleSet("default-rule-set", cmd)
	if err != nil {
		return err
	}

	overrideRuleSet, err := ReadRuleSet("override-rule-set", cmd)
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{
		Compatibility:      compatibility,
		CompatibilityGroup: compatibilityGroup,
		DefaultMetadata:    *defaultMetadata,
		OverrideMetadata:   *overrideMetadata,
		DefaultRuleSet:     *defaultRuleSet,
		OverrideRuleSet:    *overrideRuleSet,
	}

	if _, httpResp, err := srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, subject, updateReq); err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	output.Printf(errors.UpdatedSubjectLevelCompatibilityMsg, compatibility, subject)
	return nil
}

func (c *command) updateMode(subject, mode string, srClient *srsdk.APIClient, ctx context.Context) error {
	updatedMode, httpResp, err := srClient.DefaultApi.UpdateMode(ctx, subject, srsdk.ModeUpdateRequest{Mode: strings.ToUpper(mode)})
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	output.Printf(errors.UpdatedSubjectLevelModeMsg, updatedMode, subject)
	return nil
}
