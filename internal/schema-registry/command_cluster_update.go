package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
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
				Text: `Update the top-level compatibility of Schema Registry and set the compatibility group to "application.version".`,
				Code: "confluent schema-registry cluster update --compatibility backward --compatibility-group application.version",
			},
			examples.Example{
				Text: "Update top-level mode of Schema Registry.",
				Code: "confluent schema-registry cluster update --mode readwrite",
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
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	addSchemaRegistryEndpointFlag(cmd)

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
		return c.updateTopLevelMode(cmd, mode)
	}

	return fmt.Errorf(errors.CompatibilityOrModeErrorMsg)
}

func (c *command) updateTopLevelCompatibility(cmd *cobra.Command) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	req, err := c.getConfigUpdateRequest(cmd)
	if err != nil {
		return err
	}

	req, err = client.UpdateTopLevelConfig(req)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Successfully updated top-level compatibility to \"%s\".\n", req.GetCompatibility())
	return nil
}

func (c *command) getConfigUpdateRequest(cmd *cobra.Command) (srsdk.ConfigUpdateRequest, error) {
	compatibility, err := cmd.Flags().GetString("compatibility")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}

	compatibilityGroup, err := cmd.Flags().GetString("compatibility-group")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}

	req := srsdk.ConfigUpdateRequest{
		Compatibility:      srsdk.PtrString(strings.ToUpper(compatibility)),
		CompatibilityGroup: srsdk.PtrString(compatibilityGroup),
	}

	metadataDefaults, err := cmd.Flags().GetString("metadata-defaults")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}
	if metadataDefaults != "" {
		if err := read(metadataDefaults, &req.DefaultMetadata); err != nil {
			return srsdk.ConfigUpdateRequest{}, err
		}
	}

	metadataOverrides, err := cmd.Flags().GetString("metadata-overrides")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}
	if metadataOverrides != "" {
		if err := read(metadataOverrides, &req.OverrideMetadata); err != nil {
			return srsdk.ConfigUpdateRequest{}, err
		}
	}

	rulesetDefaults, err := cmd.Flags().GetString("ruleset-defaults")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}
	if rulesetDefaults != "" {
		if err := read(rulesetDefaults, &req.DefaultRuleSet); err != nil {
			return srsdk.ConfigUpdateRequest{}, err
		}
	}

	rulesetOverrides, err := cmd.Flags().GetString("ruleset-overrides")
	if err != nil {
		return srsdk.ConfigUpdateRequest{}, err
	}
	if rulesetOverrides != "" {
		if err := read(rulesetOverrides, &req.OverrideRuleSet); err != nil {
			return srsdk.ConfigUpdateRequest{}, err
		}
	}

	return req, nil
}

func (c *command) updateTopLevelMode(cmd *cobra.Command, mode string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	req := srsdk.ModeUpdateRequest{Mode: srsdk.PtrString(strings.ToUpper(mode))}

	req, err = client.UpdateTopLevelMode(req)
	if err != nil {
		return err
	}

	output.Printf(c.Config.EnableColor, "Successfully updated top-level mode to \"%s\".\n", req.GetMode())
	return nil
}
