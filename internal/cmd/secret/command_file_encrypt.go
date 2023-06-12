package secret

import (
	"github.com/spf13/cobra"
)

func (c *command) newEncryptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secrets in a configuration properties file.",
		Long:  "This command encrypts the passwords in the file specified by `--config-file`. " + masterKeyNotSetWarning,
		Args:  cobra.NoArgs,
		RunE:  c.encrypt,
	}

	c.addConfigFileFlag(cmd)
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of configuration keys.")

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("remote-secrets-file"))

	return cmd
}

func (c *command) encrypt(cmd *cobra.Command, _ []string) error {
	configs, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	configPath, localSecretsPath, remoteSecretsPath, err := c.getConfigFilePath(cmd)
	if err != nil {
		return err
	}

	return c.plugin.EncryptConfigFileSecrets(configPath, localSecretsPath, remoteSecretsPath, configs)
}
