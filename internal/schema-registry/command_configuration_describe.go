package schemaregistry

import (
	"encoding/json"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newConfigurationDescribeCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe top-level or subject-level schema configuration.",
		Args:  cobra.NoArgs,
		RunE:  c.configDescribe,
	}

	example1 := examples.Example{
		Text: `Describe the configuration of subject "payments".`,
		Code: "confluent schema-registry configuration describe --subject payments",
	}
	example2 := examples.Example{
		Text: "Describe the top-level configuration.",
		Code: "confluent schema-registry configuration describe",
	}
	if cfg.IsOnPremLogin() {
		example1.Code += " " + onPremAuthenticationMsg
		example2.Code += " " + onPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("subject", "", subjectUsage)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		addCaLocationFlag(cmd)
		addSchemaRegistryEndpointFlag(cmd)
	}
	pcmd.AddOutputFlag(cmd)

	if cfg.IsCloudLogin() {
		// Deprecated
		pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-key"))

		// Deprecated
		pcmd.AddApiSecretFlag(cmd)
		cobra.CheckErr(cmd.Flags().MarkHidden("api-secret"))
	}

	return cmd
}

func (c *command) configDescribe(cmd *cobra.Command, _ []string) error {
	client, err := c.GetSchemaRegistryClient(cmd)
	if err != nil {
		return err
	}

	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	var config srsdk.Config
	if subject != "" {
		config, err = client.GetSubjectLevelConfig(subject)
		if err != nil {
			return catchSubjectLevelConfigNotFoundError(err, subject)
		}
	} else {
		config, err = client.GetTopLevelConfig()
		if err != nil {
			return err
		}
	}

	out := &configOut{
		CompatibilityLevel: config.GetCompatibilityLevel(),
		CompatibilityGroup: config.GetCompatibilityGroup(),
	}

	if config.DefaultMetadata.IsSet() {
		defaultMetadata, err := json.Marshal(config.DefaultMetadata)
		if err != nil {
			return err
		}
		out.MetadataDefaults = prettyJson(defaultMetadata)
	}

	if config.OverrideMetadata.IsSet() {
		overrideMetadata, err := json.Marshal(config.OverrideMetadata)
		if err != nil {
			return err
		}
		out.MetadataOverrides = prettyJson(overrideMetadata)
	}

	if config.DefaultRuleSet.IsSet() {
		defaultRuleset, err := json.Marshal(config.DefaultRuleSet)
		if err != nil {
			return err
		}
		out.RulesetDefaults = prettyJson(defaultRuleset)
	}

	if config.OverrideRuleSet.IsSet() {
		overrideRuleset, err := json.Marshal(config.OverrideRuleSet)
		if err != nil {
			return err
		}
		out.RulesetOverrides = prettyJson(overrideRuleset)
	}

	table := output.NewTable(cmd)
	table.Add(out)
	return table.PrintWithAutoWrap(false)
}
