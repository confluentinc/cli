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
	addCompatibilityGroupFlag(cmd)
	addMetadataDefaultsFlag(cmd)
	addMetadataOverridesFlag(cmd)
	addRulesetDefaultsFlag(cmd)
	addRulesetOverridesFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) subjectUpdate(cmd *cobra.Command, args []string) error {
	subject := args[0]

	srClient, ctx, err := getApiClient(cmd, c.Config, c.Version)
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
		return c.updateMode(subject, mode, srClient, ctx)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *command) updateCompatibility(cmd *cobra.Command, subject, compatibility string, srClient *srsdk.APIClient, ctx context.Context) error {
	updateReq, err := c.getConfigUpdateRequest(cmd)
	if err != nil {
		return err
	}

	if _, httpResp, err := srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, subject, updateReq); err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	output.Printf("Successfully updated subject level compatibility to \"%s\" for subject \"%s\".\n", compatibility, subject)
	return nil
}

func (c *command) updateMode(subject, mode string, srClient *srsdk.APIClient, ctx context.Context) error {
	updatedMode, httpResp, err := srClient.DefaultApi.UpdateMode(ctx, subject, srsdk.ModeUpdateRequest{Mode: strings.ToUpper(mode)})
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	output.Printf("Successfully updated subject level mode to \"%s\" for subject \"%s\".\n", updatedMode, subject)
	return nil
}
