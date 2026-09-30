package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v4/pkg/config"
)

type linkConfigurationOut struct {
	ConfigName  string   `human:"Config Name" serialized:"config_name"`
	ConfigValue string   `human:"Config Value" serialized:"config_value"`
	ReadOnly    bool     `human:"Read-Only" serialized:"read_only"`
	Sensitive   bool     `human:"Sensitive" serialized:"sensitive"`
	Source      string   `human:"Source" serialized:"source"`
	Synonyms    []string `human:"Synonyms" serialized:"synonyms"`
}

func (c *linkCommand) newConfigurationCommand(cfg *config.Config) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "configuration",
		Short: "Manage cluster link configurations.",
	}

	if cfg.IsCloudLogin() {
		cmd.AddCommand(c.newConfigurationListCommand())
		cmd.AddCommand(c.newConfigurationUpdateCommand())
	} else {
		cmd.AddCommand(c.newConfigurationListCommandOnPrem())
		cmd.AddCommand(c.newConfigurationUpdateCommandOnPrem())
	}

	return cmd
}
