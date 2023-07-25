package schemaregistry

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type configOut struct {
	CompatibilityLevel string `human:"Compatibility Level,omitempty" serialized:"compatibility_level,omitempty"`
	CompatibilityGroup string `human:"Compatibility Group,omitempty" serialized:"compatibility_group,omitempty"`
	MetadataDefaults   string `human:"Metadata Defaults,omitempty" serialized:"metadata_defaults,omitempty"`
	MetadataOverrides  string `human:"Metadata Overrides,omitempty" serialized:"metadata_overrides,omitempty"`
	RulesetDefaults    string `human:"Ruleset Defaults,omitempty" serialized:"ruleset_defaults,omitempty"`
	RulesetOverrides   string `human:"Ruleset Overrides,omitempty" serialized:"ruleset_overrides,omitempty"`
}

func (c *command) newConfigDescribeCommand(cfg *v1.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe top-level or subject-level schema configuration.",
		Args:  cobra.NoArgs,
		RunE:  c.configDescribe,
	}

	example1 := examples.Example{
		Text: `Describe the configuration of subject "payments".`,
		Code: "confluent schema-registry config describe --subject payments",
	}
	example2 := examples.Example{
		Text: "Describe the top-level configuration.",
		Code: "confluent schema-registry config describe",
	}
	if !cfg.IsCloudLogin() {
		example1.Code += " " + OnPremAuthenticationMsg
		example2.Code += " " + OnPremAuthenticationMsg
	}
	cmd.Example = examples.BuildExampleString(example1, example2)

	cmd.Flags().String("subject", "", SubjectUsage)
	if cfg.IsCloudLogin() {
		pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	} else {
		cmd.Flags().AddFlagSet(pcmd.OnPremSchemaRegistrySet())
	}
	pcmd.AddContextFlag(cmd, c.CLICommand)
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

func (c *command) configDescribe(cmd *cobra.Command, args []string) error {
	client, err := c.GetSchemaRegistryClient()
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
		CompatibilityLevel: config.CompatibilityLevel,
		CompatibilityGroup: config.CompatibilityGroup,
	}

	if config.DefaultMetadata != nil {
		defaultMetadata, err := json.Marshal(config.DefaultMetadata)
		if err != nil {
			return err
		}
		out.MetadataDefaults = prettyJson(defaultMetadata)
	}

	if config.OverrideMetadata != nil {
		overrideMetadata, err := json.Marshal(config.OverrideMetadata)
		if err != nil {
			return err
		}
		out.MetadataOverrides = prettyJson(overrideMetadata)
	}

	if config.DefaultRuleSet != nil {
		defaultRuleset, err := json.Marshal(config.DefaultRuleSet)
		if err != nil {
			return err
		}
		out.RulesetDefaults = prettyJson(defaultRuleset)
	}

	if config.OverrideRuleSet != nil {
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

func catchSubjectLevelConfigNotFoundError(err error, subject string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return errors.New(fmt.Sprintf(`subject "%s" does not have subject-level compatibility configured`, subject))
	}

	return err
}

func prettyJson(str []byte) string {
	return strings.TrimSpace(string(pretty.Pretty(str)))
}
