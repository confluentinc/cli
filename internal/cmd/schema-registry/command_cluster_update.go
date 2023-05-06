package schemaregistry

import (
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

func (c *command) newClusterUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:         "update",
		Short:       "Update global mode or compatibility of Schema Registry.",
		Args:        cobra.NoArgs,
		RunE:        c.clusterUpdate,
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLogin},
		Example: examples.BuildExampleString(
			examples.Example{
				Text: "Update top-level compatibility of Schema Registry.",
				Code: "confluent schema-registry cluster update --compatibility backward",
			},
			examples.Example{
				Text: `Update top-level compatibility of Schema Registry and set compatibility group to "application.version".`,
				Code: "confluent schema-registry cluster update --compatibility backward --compatibility-group application.version",
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
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)

	return cmd
}

func (c *command) clusterUpdate(cmd *cobra.Command, _ []string) error {
	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}
	if compatibility != "" {
		return c.updateTopLevelCompatibility(cmd)
	}

	mode, err := cmd.Flags().GetString("mode")
	if err != nil {
		return err
	}
	if mode != "" {
		return c.updateTopLevelMode(cmd)
	}

	return errors.New(errors.CompatibilityOrModeErrorMsg)
}

func (c *command) updateTopLevelCompatibility(cmd *cobra.Command) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return err
	}

	compatibilityGroup, err := cmd.Flags().GetString("compatibility-group")
	if err != nil {
		return err
	}

	var defaultMetadata srsdk.Metadata
	var overrideMetadata srsdk.Metadata
	var defaultRuleset srsdk.RuleSet
	var overrideRuleset srsdk.RuleSet

	metadataDefaultsPath, err := cmd.Flags().GetString("metadata-defaults")
	if err != nil {
		return err
	}
	rulesetDefaultPath, err := cmd.Flags().GetString("ruleset-defaults")
	if err != nil {
		return err
	}
	err = readMetadataAndRuleset(metadataDefaultsPath, &defaultMetadata, rulesetDefaultPath, &defaultRuleset)
	if err != nil {
		return err
	}

	metadataOverridesPath, err := cmd.Flags().GetString("metadata-overrides")
	if err != nil {
		return err
	}
	rulesetOverridesPath, err := cmd.Flags().GetString("ruleset-overrides")
	if err != nil {
		return err
	}
	err = readMetadataAndRuleset(metadataOverridesPath, &overrideMetadata, rulesetOverridesPath, &overrideRuleset)
	if err != nil {
		return err
	}

	updateReq := srsdk.ConfigUpdateRequest{
		Compatibility:      strings.ToUpper(compatibility),
		CompatibilityGroup: compatibilityGroup,
		DefaultMetadata:    defaultMetadata,
		OverrideMetadata:   overrideMetadata,
		DefaultRuleSet:     defaultRuleset,
		OverrideRuleSet:    overrideRuleset,
	}

	if _, _, err := srClient.DefaultApi.UpdateTopLevelConfig(ctx, updateReq); err != nil {
		return err
	}

	output.Printf(errors.UpdatedToLevelCompatibilityMsg, updateReq.Compatibility)
	return nil
}

func (c *command) updateTopLevelMode(cmd *cobra.Command) error {
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

	output.Printf(errors.UpdatedTopLevelModeMsg, modeUpdate.Mode)
	return nil
}
