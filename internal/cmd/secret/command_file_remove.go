package secret

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove values from a configuration properties file.",
		Args:  cobra.NoArgs,
		RunE:  c.remove,
	}

	cmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of configuration keys.")

	_ = cmd.MarkFlagRequired("config-file")
	_ = cmd.MarkFlagRequired("local-secrets-file")
	_ = cmd.MarkFlagRequired("config")

	return cmd
}

func (c *command) remove(cmd *cobra.Command, _ []string) error {
	configSource, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	removeConfigs, err := c.getConfigs(configSource, "config properties", "", false)
	if err != nil {
		return err
	}

	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	if err := c.plugin.RemoveEncryptedPasswords(configPath, localSecretsPath, removeConfigs); err != nil {
		return err
	}

	utils.ErrPrintln(cmd, "Deleted configuration values.")
	return nil
}
