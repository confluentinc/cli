package schemaregistry

import (
	"context"
	"fmt"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func (c *subjectCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update <subject>",
		Short: "Update subject compatibility or mode.",
		Args:  cobra.ExactArgs(1),
		RunE:  pcmd.NewCLIRunE(c.update),
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Update subject level compatibility of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --compatibility=BACKWARD", version.CLIName),
			},
			examples.Example{
				Text: `Update subject level mode of subject "payments".`,
				Code: fmt.Sprintf("%s schema-registry subject update payments --mode=READWRITE", version.CLIName),
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

func (c *subjectCommand) update(cmd *cobra.Command, args []string) error {
	subject := args[0]

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
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

func (c *subjectCommand) updateCompatibility(cmd *cobra.Command, subject, compatibility string, srClient *srsdk.APIClient, ctx context.Context) error {
	updateReq := srsdk.ConfigUpdateRequest{Compatibility: compatibility}
	if _, httpResp, err := srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, subject, updateReq); err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelCompatibilityMsg, compatibility, subject)
	return nil
}

func (c *subjectCommand) updateMode(cmd *cobra.Command, subject, mode string, srClient *srsdk.APIClient, ctx context.Context) error {
	updatedMode, httpResp, err := srClient.DefaultApi.UpdateMode(ctx, subject, srsdk.ModeUpdateRequest{Mode: mode})
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, httpResp)
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelModeMsg, updatedMode, subject)
	return nil
}
