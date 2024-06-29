package kafka

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/config"
)

type linkConfigurationHumanOut struct {
	ConfigName  string `human:"Config Name"`
	ConfigValue string `human:"Config Value"`
	ReadOnly    bool   `human:"Read-Only"`
	Sensitive   bool   `human:"Sensitive"`
	Source      string `human:"Source"`
	Synonyms    string `human:"Synonyms"`
}

type linkConfigurationSerializedOut struct {
	ConfigName  string   `serialized:"config_name"`
	ConfigValue string   `serialized:"config_value"`
	ReadOnly    bool     `serialized:"read_only"`
	Sensitive   bool     `serialized:"sensitive"`
	Source      string   `serialized:"source"`
	Synonyms    []string `serialized:"synonyms"`
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
