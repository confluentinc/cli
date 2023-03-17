package secret

import (
	"github.com/spf13/cobra"
)

func (c *command) newAddCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add secrets to a configuration properties file.",
		Long:  "This command encrypts the password and adds it to the configuration file specified by `--config-file`. " + masterKeyNotSetWarning,
		Args:  cobra.NoArgs,
		RunE:  c.add,
	}

	cmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of key/value pairs of configuration properties.")

	cobra.CheckErr(cmd.MarkFlagFilename("config-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagFilename("remote-secrets-file"))

	cobra.CheckErr(cmd.MarkFlagRequired("config-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("remote-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("config"))

	return cmd
}

func (c *command) add(cmd *cobra.Command, _ []string) error {
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

	return c.plugin.AddEncryptedPasswords(configPath, localSecretsPath, remoteSecretsPath, newConfigs)
}
