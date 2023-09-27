package secret

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newRemoveCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove values from a configuration properties file.",
		Args:  cobra.NoArgs,
		RunE:  c.remove,
	}

	c.addConfigFileFlag(cmd)
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of configuration keys.")

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *command) remove(cmd *cobra.Command, _ []string) error {
	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	removeConfigs, err := c.getConfigs(config, "config properties", "", false)
	if err != nil {
		return err
	}

	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	localSecretsFile, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	if err := c.plugin.RemoveEncryptedPasswords(configFile, localSecretsFile, removeConfigs); err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, "Deleted configuration values.")
	return nil
}
