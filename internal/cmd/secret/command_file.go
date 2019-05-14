package secret

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	secret "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"bufio"
	"strings"
)

type secureFileCommand struct {
	*cobra.Command
	config *config.Config
	plugin secret.PasswordProtection
}

// NewFileCommand returns the Cobra command for managing encrypted file.
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
		Args:  cobra.NoArgs,
	}
	encryptCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = encryptCmd.MarkFlagRequired("config-file-path")

	encryptCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = encryptCmd.MarkFlagRequired("local-secrets-file-path")

	encryptCmd.Flags().String("remote-secrets-file-path", "", "Remote Encrypted Config Properties File Path.")
	_ = encryptCmd.MarkFlagRequired("remote-secrets-file-path")
	c.AddCommand(encryptCmd)

	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt encrypted secrets from config properties file.",
		RunE:  c.encrypt,
		Args:  cobra.NoArgs,
	}
	decryptCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = decryptCmd.MarkFlagRequired("config-file-path")

	decryptCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = decryptCmd.MarkFlagRequired("local-secrets-file-path")

	decryptCmd.Flags().String("output-file-path", "", "Output File Path.")
	_ = decryptCmd.MarkFlagRequired("output-file-path")
	c.AddCommand(decryptCmd)

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add encrypted secrets to a config properties file.",
		RunE:  c.add,
		Args:  cobra.NoArgs,
	}
	addCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = addCmd.MarkFlagRequired("config-file-path")

	addCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = addCmd.MarkFlagRequired("local-secrets-file-path")

	addCmd.Flags().String("remote-secrets-file-path", "", "Remote Encrypted Config Properties File Path.")
	_ = addCmd.MarkFlagRequired("remote-secrets-file-path")

	addCmd.Flags().String("config", "", "List of config properties")
	_ = addCmd.MarkFlagRequired("config")
	c.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update encrypted secrets from config properties file.",
		RunE:  c.add,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = updateCmd.MarkFlagRequired("config-file-path")

	updateCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = updateCmd.MarkFlagRequired("local-secrets-file-path")

	updateCmd.Flags().String("remote-secrets-file-path", "", "Remote Encrypted Config Properties File Path.")
	_ = updateCmd.MarkFlagRequired("remote-secrets-file-path")

	updateCmd.Flags().String("config", "", "List of config properties")
	_ = updateCmd.MarkFlagRequired("config")
	c.AddCommand(updateCmd)

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete configs from config properties file.",
		RunE:  c.remove,
		Args:  cobra.NoArgs,
	}
	removeCmd.Flags().String("config-file-path", "", "Config Properties File Path.")
	_ = removeCmd.MarkFlagRequired("config-file-path")

	removeCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = removeCmd.MarkFlagRequired("local-secrets-file-path")

	removeCmd.Flags().String("config", "", "List of config properties")
	_ = removeCmd.MarkFlagRequired("config")
	c.AddCommand(removeCmd)
}

func (c *secureFileCommand) encrypt(cmd *cobra.Command, args []string) error {
	configPath, localSecretsPath, remoteSecretsPath, err := c.getConfigFilePath(cmd)
	if err != nil {
		return err
	}

	err = c.plugin.EncryptConfigFileSecrets(configPath, localSecretsPath, remoteSecretsPath)
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

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputPath, err := cmd.Flags().GetString("output-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.DecryptConfigFileSecrets(configPath, localSecretsPath, outputPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) getConfigs(configSource string) (string, error) {
	if configSource == "" {
		return "", fmt.Errorf("Please enter config properties.")
	}

	if configSource == "-" {
		reader := bufio.NewReader(os.Stdin)
		configs, err := reader.ReadString('\n')
		return configs, err
	}

	if strings.HasPrefix(configSource, "@") {
		filePath := configSource[1:len(configSource)]
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		configs := string(data)
		return configs, err
	}

	return "", fmt.Errorf("Invalid config properties source")
}

func (c *secureFileCommand) add(cmd *cobra.Command, args []string) error {
	configSource,  err :=  cmd.Flags().GetString("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	newConfigs, err := c.getConfigs(configSource)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	configPath, localSecretsPath, remoteSecretsPath, err := c.getConfigFilePath(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.AddEncryptedPasswords(configPath, localSecretsPath, remoteSecretsPath, newConfigs)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) getConfigFilePath(cmd *cobra.Command) (string, string, string, error) {
	configPath, err := cmd.Flags().GetString("config-file-path")
	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file-path")
	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}

	remoteSecretsPath, err := cmd.Flags().GetString("remote-secrets-file-path")
	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}

	return configPath, localSecretsPath, remoteSecretsPath, nil
}

func (c *secureFileCommand) remove(cmd *cobra.Command, args []string) error {
	configSource, err := cmd.Flags().GetString("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	removeConfigs, err := c.getConfigs(configSource)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	configPath, err := cmd.Flags().GetString("config-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.RemoveEncryptedPasswords(configPath, localSecretsPath, removeConfigs)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}
