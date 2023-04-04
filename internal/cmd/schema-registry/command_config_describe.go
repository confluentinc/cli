package schemaregistry

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/spf13/cobra"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/examples"
	"github.com/confluentinc/cli/internal/pkg/output"
)

type configOut struct {
	CompatibilityLevel string `human:"Compatibility Level,omitempty" serialized:"compatibility_level,omitempty"`
	CompatibilityGroup string `human:"Compatibility Group,omitempty" serialized:"compatibility_group,omitempty"`
	DefaultMetadata    string `human:"Default Metadata,omitempty" serialized:"default_metadata,omitempty"`
	OverrideMetadata   string `human:"Overridden Metadata,omitempty" serialized:"overridden_metadata,omitempty"`
	DefaultRuleSet     string `human:"Default Ruleset,omitempty" serialized:"default_ruleset,omitempty"`
	OverrideRuleSet    string `human:"Overridden Ruleset,omitempty" serialized:"overridden_ruleset,omitempty"`
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

	defaultMetadata, err := json.Marshal(config.DefaultMetadata)
	if err != nil {
		return err
	}

	overrideMetadata, err := json.Marshal(config.OverrideMetadata)
	if err != nil {
		return err
	}

	defaultRuleSet, err := json.Marshal(config.DefaultRuleSet)
	if err != nil {
		return err
	}

	overrideRuleSet, err := json.Marshal(config.OverrideRuleSet)
	if err != nil {
		return err
	}

	table := output.NewTable(cmd)
	table.Add(&configOut{CompatibilityLevel: config.CompatibilityLevel,
		CompatibilityGroup: config.CompatibilityGroup,
		DefaultMetadata:    string(defaultMetadata),
		OverrideMetadata:   string(overrideMetadata),
		DefaultRuleSet:     string(defaultRuleSet),
		OverrideRuleSet:    string(overrideRuleSet),
	})
	return table.Print()
}
