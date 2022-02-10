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
				Text: "Update subject level compatibility or mode of Schema Registry:",
				Code: fmt.Sprintf("%s schema-registry subject update payments --compatibility=BACKWARD\n%s schema-registry subject update payments --mode=READWRITE", version.CLIName, version.CLIName),
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
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}
	return c.updateSchemaSubject(cmd, args, srClient, ctx)
}

func (c *subjectCommand) updateSchemaSubject(cmd *cobra.Command, args []string, srClient *srsdk.APIClient, ctx context.Context) error {
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

func (c *subjectCommand) updateCompatibility(cmd *cobra.Command, args []string, srClient *srsdk.APIClient, ctx context.Context) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: compat}
	if _, r, err := srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, args[0], updateReq); err != nil {
		return errors.CatchSchemaNotFoundError(err, r)
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelCompatibilityMsg, compat, args[0])
	return nil
}

func (c *subjectCommand) updateMode(cmd *cobra.Command, args []string, srClient *srsdk.APIClient, ctx context.Context) error {
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	updatedMode, r, err := srClient.DefaultApi.UpdateMode(ctx, args[0], srsdk.ModeUpdateRequest{Mode: mode})
	if err != nil {
		return errors.CatchSchemaNotFoundError(err, r)
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelModeMsg, updatedMode, args[0])
	return nil
}
