package secret

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	secret "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/spf13/cobra"
)

type secureFileCommand struct {
	*cobra.Command
	config *config.Config
	plugin secret.PasswordProtection
}

// NewFileCommand returns the Cobra command for managing master key.
func NewFileCommand(config *config.Config, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &secureFileCommand{
		Command: &cobra.Command{
			Use:   "file",
			Short: "Secure secrets in Config File",
		},
		config: config,
		plugin: plugin,
	}
	cmd.init()
	return cmd.Command
}

func (c *secureFileCommand) init() {
	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secrets in config properties file.",
		RunE:  c.encrypt,
		Args:  cobra.ExactArgs(1),
	}
	encryptCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = encryptCmd.MarkFlagRequired("config-file-path")
	c.AddCommand(encryptCmd)

	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt encrypted secrets from config properties file.",
		RunE:  c.encrypt,
		Args:  cobra.ExactArgs(1),
	}
	decryptCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = decryptCmd.MarkFlagRequired("config-file-path")
	decryptCmd.Flags().String("output-file-path", "", "Output File Path.")
	_ = decryptCmd.MarkFlagRequired("output-file-path")
	c.AddCommand(decryptCmd)

	editCmd := &cobra.Command{
		Use:   "edit",
		Short: "Add encrypted secrets to a config properties file.",
		RunE:  c.edit,
		Args:  cobra.ExactArgs(1),
	}
	editCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = editCmd.MarkFlagRequired("config-file-path")
	c.AddCommand(editCmd)
}

func (c *secureFileCommand) encrypt(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString("config-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.EncryptConfigFileSecrets(configPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) decrypt(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString("config-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputPath, err := cmd.Flags().GetString("output-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.DecryptConfigFileSecrets(configPath, outputPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) edit(cmd *cobra.Command, args []string) error {
	configPath, err := cmd.Flags().GetString("config-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.AddEncryptedPasswords(configPath, configPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}
