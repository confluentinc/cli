package schemaregistry

import (
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *clusterCommand) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       "Update global mode or compatibility of Schema Registry.",
		Args:        cobra.NoArgs,
		RunE:        pcmd.NewCLIRunE(c.update),
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update top level compatibility or mode of Schema Registry.",
				Code: "confluent schema-registry cluster update --compatibility=BACKWARD",
			},
			examples.Example{
				Code: "confluent schema-registry cluster update --mode=READWRITE",
			},
		),
	}

	addCompatibilityFlag(cmd)
	addModeFlag(cmd)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)

	return cmd
}

func (c *clusterCommand) update(cmd *cobra.Command, _ []string) error {
	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compat != "" {
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
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	compat, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{Compatibility: strings.ToUpper(compat)}

	if _, _, err := srClient.DefaultApi.UpdateTopLevelConfig(ctx, updateReq); err != nil {
		return err
	}

	utils.Printf(cmd, errors.UpdatedToLevelCompatibilityMsg, updateReq.Compatibility)
	return nil
}

func (c *clusterCommand) updateMode(cmd *cobra.Command) error {
	srClient, ctx, err := GetApiClient(cmd, c.srClient, c.Config, c.Version)
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

	utils.Printf(cmd, errors.UpdatedTopLevelModeMsg, modeUpdate.Mode)
	return nil
}
