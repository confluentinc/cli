package secret

import (
	"github.com/spf13/cobra"
)

func (c *command) newEncryptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secrets in a configuration properties file.",
		Long:  "This command encrypts the passwords in file specified in --config-file. This command returns a failure if a master key has not already been set in the environment variable. Create master key using \"master-key generate\" command and save the generated master key in environment variable.",
		Args:  cobra.NoArgs,
		RunE:  c.encrypt,
	}

	cmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of configuration keys.")

	_ = cmd.MarkFlagRequired("config-file")
	_ = cmd.MarkFlagRequired("local-secrets-file")
	_ = cmd.MarkFlagRequired("remote-secrets-file")

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

	cipherMode := c.getCipherMode()
	c.plugin.SetCipherMode(cipherMode)

	return c.plugin.EncryptConfigFileSecrets(configPath, localSecretsPath, remoteSecretsPath, configs)
}
