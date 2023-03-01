package schemaregistry

import (
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newUpdateCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       "Update global mode or compatibility of Schema Registry.",
		Args:        cobra.NoArgs,
		RunE:        c.update,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update top-level compatibility of Schema Registry.",
				Code: "confluent schema-registry cluster update --compatibility backward",
			},
			examples.Example{
				Text: "Update top-level mode of Schema Registry.",
				Code: "confluent schema-registry cluster update --mode readwrite",
			},
		),
	}

	addCompatibilityFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	}

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, _ []string) error {
	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compatibility != "" {
		return c.updateCompatibility(cmd)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateMode(cmd)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *clusterCommand) updateCompatibility(cmd *cobra.Command) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: strings.ToUpper(compatibility)}

	if _, _, err := srClient.DefaultApi.UpdateTopLevelConfig(ctx, updateReq); err != nil {
		return err
	}

	utils.Printf(errors.UpdatedToLevelCompatibilityMsg, updateReq.Compatibility)
	return nil
}

func (c *clusterCommand) updateMode(cmd *cobra.Command) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}

	modeUpdate, _, err := srClient.DefaultApi.UpdateTopLevelMode(ctx, srsdk.ModeUpdateRequest{Mode: strings.ToUpper(mode)})
	if err != nil {
		return err
	}

	utils.Printf(errors.UpdatedTopLevelModeMsg, modeUpdate.Mode)
	return nil
}
