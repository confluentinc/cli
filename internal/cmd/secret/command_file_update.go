package secret

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newUpdateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update secrets in a configuration properties file.",
		Long:  "This command updates the encrypted secrets from the configuration properties file.",
		Args:  cobra.NoArgs,
		RunE:  c.update,
	}

	cmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	cmd.Flags().String("config", "", "List of key/value pairs of configuration properties.")

	_ = cmd.MarkFlagRequired("config-file")
	_ = cmd.MarkFlagRequired("local-secrets-file")
	_ = cmd.MarkFlagRequired("remote-secrets-file")
	_ = cmd.MarkFlagRequired("config")

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

	utils.ErrPrintln(cmd, errors.UpdateSecretFileMsg)
	return nil
}
