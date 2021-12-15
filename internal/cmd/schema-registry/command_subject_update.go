package schemaregistry

import (
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
				Code: fmt.Sprintf("%s schema-registry subject update <subject-name> --compatibility=BACKWARD\n%s schema-registry subject update <subject-name> --mode=READWRITE", version.CLIName, version.CLIName),
			},
		),
	}

	addCompatibilityFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)

	return cmd
}

func (c *subjectCommand) update(cmd *cobra.Command, args []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compat != "" {
		return c.updateCompatibility(cmd, args)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateMode(cmd, args)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *subjectCommand) updateCompatibility(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: compat}
	if _, _, err = srClient.DefaultApi.UpdateSubjectLevelConfig(ctx, args[0], updateReq); err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelCompatibilityMsg, compat, args[0])
	return nil
}

func (c *subjectCommand) updateMode(cmd *cobra.Command, args []string) error {
	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	updatedMode, _, err := srClient.DefaultApi.UpdateMode(ctx, args[0], srsdk.ModeUpdateRequest{Mode: mode})
	if err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedSubjectLevelModeMsg, updatedMode, args[0])
	return nil
}
