package schemaregistry

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tidwall/pretty"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
)

type configOut struct {
	CompatibilityLevel string `human:"Compatibility Level,omitempty" json:"compatibilityLevel" serialized:"compatibility_level,omitempty"`
	CompatibilityGroup string `human:"Compatibility Group,omitempty" json:"compatibilityGroup" serialized:"compatibility_group,omitempty"`
	MetadataDefaults   string `human:"Metadata Defaults,omitempty" json:"metadataDefaults" serialized:"metadata_defaults,omitempty"`
	MetadataOverrides  string `human:"Metadata Overrides,omitempty" json:"metadataOverrides" serialized:"metadata_overrides,omitempty"`
	RulesetDefaults    string `human:"Ruleset Defaults,omitempty" json:"rulesetDefaults" serialized:"ruleset_defaults,omitempty"`
	RulesetOverrides   string `human:"Ruleset Overrides,omitempty" json:"rulesetOverrides" serialized:"ruleset_overrides,omitempty"`
}

func (c *command) newConfigCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:         "config",
		Short:       "Manage Schema Registry configuration.",
		Annotations: map[string]string{pcmd.RunRequirement: pcmd.RequireCloudLoginOrOnPremLogin},
	}

	cmd.AddCommand(c.newConfigDeleteCommand(cfg))
	cmd.AddCommand(c.newConfigDescribeCommand(cfg))

	return cmd
}

func catchSubjectLevelConfigNotFoundError(err error, subject string) error {
	if err != nil && strings.Contains(err.Error(), "Not Found") {
		return fmt.Errorf(fmt.Sprintf(`subject "%s" does not have subject-level compatibility configured`, subject))
	}

	return err
}

func prettyJson(str []byte) string {
	return strings.TrimSpace(string(pretty.Pretty(str)))
}
