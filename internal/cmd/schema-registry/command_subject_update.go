package schemaregistry

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
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
				Code: "confluent schema-registry subject update payments --compatibility backward",
			},
			examples.Example{
				Text: `Update subject-level compatibility of subject "payments" and set compatibility group to "application.version".`,
				Code: "confluent schema-registry subject update payments --compatibility backward --compatibility-group application.version",
			},
			examples.Example{
				Text: `Update subject-level mode of subject "payments".`,
				Code: "confluent schema-registry subject update payments --mode readwrite",
			},
		),
	}

	addCompatibilityFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

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

	metadataPath, err := cmd.Flags().GetString("metadata-defaults")
	if err != nil {
		return err
	}
	defaultMetadata, err := readMetadata(metadataPath)
	if err != nil {
		return err
	}

	metadataPath, err = cmd.Flags().GetString("metadata-overrides")
	if err != nil {
		return err
	}
	overrideMetadata, err := readMetadata(metadataPath)
	if err != nil {
		return err
	}

	rulesetPath, err := cmd.Flags().GetString("ruleset-defaults")
	if err != nil {
		return err
	}
	defaultRuleSet, err := readRuleset(rulesetPath)
	if err != nil {
		return err
	}

	rulesetPath, err = cmd.Flags().GetString("ruleset-overrides")
	if err != nil {
		return err
	}
	overrideRuleSet, err := readRuleset(rulesetPath)
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
