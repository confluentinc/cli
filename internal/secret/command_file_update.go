package secret

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update secrets in a configuration properties file.",
		Long:  "This command updates the encrypted secrets from the configuration properties file.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	c.addConfigFileFlag(cmd)
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of key/value pairs of configuration properties.")

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("remote-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *command) update(cmd *cobra.Command, _ []string) error {
	config, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	newConfigs, err := c.getConfigs(config, "config properties", "", false)
	if err != nil {
		return err
	}

	configPath, localSecretsPath, remoteSecretsPath, err := c.getConfigFilePath(cmd)
	if err != nil {
		return err
	}

	if err := c.plugin.UpdateEncryptedPasswords(configPath, localSecretsPath, remoteSecretsPath, newConfigs); err != nil {
		return err
	}

	output.Println(c.Config.EnableColor, "Updated the encrypted secrets.")
	return nil
}
