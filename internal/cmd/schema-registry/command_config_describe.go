package schemaregistry

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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

func (c *command) newConfigDescribeCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe",
		Short: "Describe top-level or subject-level schema compatibility.",
		Args:  cobra.NoArgs,
		RunE:  c.configDescribe,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Describe the configuration of subject "payments".`,
				Code: "confluent schema-registry config describe --subject payments",
			},
			examples.Example{
				Text: "Describe the top-level configuration.",
				Code: "confluent schema-registry config describe",
			},
		),
	}

	cmd.Flags().String("subject", "", SubjectUsage)
	pcmd.AddApiKeyFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddApiSecretFlag(cmd)
	pcmd.AddContextFlag(cmd, c.CLICommand)
	pcmd.AddEnvironmentFlag(cmd, c.AuthenticatedCLICommand)
	pcmd.AddOutputFlag(cmd)

	return cmd
}

func (c *command) configDescribe(cmd *cobra.Command, args []string) error {
	srClient, ctx, err := getApiClient(cmd, c.srClient, c.Config, c.Version)
	if err != nil {
		return err
	}

	return describeSchemaConfig(cmd, srClient, ctx)
}

func describeSchemaConfig(cmd *cobra.Command, srClient *srsdk.APIClient, ctx context.Context) error {
	subject, err := cmd.Flags().GetString("subject")
	if err != nil {
		return err
	}

	var config srsdk.Config
	var httpResp *http.Response
	if subject != "" {
		config, httpResp, err = srClient.DefaultApi.GetSubjectLevelConfig(ctx, subject, nil)
		if err != nil {
			return errors.CatchNoSubjectLevelConfigError(err, httpResp, subject)
		}
	} else {
		config, _, err = srClient.DefaultApi.GetTopLevelConfig(ctx)
		if err != nil {
			return err
		}
	}

	configOut := &configOut{
		CompatibilityLevel: config.CompatibilityLevel,
		CompatibilityGroup: config.CompatibilityGroup,
	}

	if config.DefaultMetadata != nil {
		defaultMetadata, err := json.Marshal(config.DefaultMetadata)
		if err != nil {
			return err
		}
		configOut.MetadataDefaults = prettyJson(defaultMetadata)
	}

	if config.OverrideMetadata != nil {
		overrideMetadata, err := json.Marshal(config.OverrideMetadata)
		if err != nil {
			return err
		}
		configOut.MetadataOverrides = prettyJson(overrideMetadata)
	}

	if config.DefaultRuleSet != nil {
		defaultRuleset, err := json.Marshal(config.DefaultRuleSet)
		if err != nil {
			return err
		}
		configOut.RulesetDefaults = prettyJson(defaultRuleset)
	}

	if config.OverrideRuleSet != nil {
		overrideRuleset, err := json.Marshal(config.OverrideRuleSet)
		if err != nil {
			return err
		}
		configOut.RulesetOverrides = prettyJson(overrideRuleset)
	}

	table := output.NewTable(cmd)
	table.Add(configOut)
	return table.PrintWithAutoWrap(false)
}

func prettyJson(str []byte) string {
	return strings.TrimSuffix(string(pretty.Pretty(str)), "\n")
}
