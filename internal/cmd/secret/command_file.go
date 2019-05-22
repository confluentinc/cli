package secret

import (
	"bufio"
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/spf13/cobra"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"io/ioutil"
	"os"
	"strings"
)

type secureFileCommand struct {
	*cobra.Command
	config *config.Config
	plugin secret.PasswordProtection
	prompt pcmd.Prompt
}

// NewFileCommand returns the Cobra command for managing encrypted file.
func NewFileCommand(config *config.Config, prompt pcmd.Prompt, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &secureFileCommand{
		Command: &cobra.Command{
			Use:   "file",
			Short: "Secure secrets in a configuration properties file.",
		},
		config: config,
		plugin: plugin,
		prompt: prompt,
	}
	cmd.init()
	return cmd.Command
}

func (c *secureFileCommand) init() {
	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secrets in a configuration properties file.",
		RunE:  c.encrypt,
		Args:  cobra.NoArgs,
	}
	encryptCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	_ = encryptCmd.MarkFlagRequired("config-file")

	encryptCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	_ = encryptCmd.MarkFlagRequired("local-secrets-file")

	encryptCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	_ = encryptCmd.MarkFlagRequired("remote-secrets-file")
	encryptCmd.Flags().SortFlags = false
	c.AddCommand(encryptCmd)

	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt encrypted secrets from the configuration properties file",
		RunE:  c.decrypt,
		Args:  cobra.NoArgs,
	}
	decryptCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	_ = decryptCmd.MarkFlagRequired("config-file")

	decryptCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	_ = decryptCmd.MarkFlagRequired("local-secrets-file")

	decryptCmd.Flags().String("output-file", "", "Output file path.")
	_ = decryptCmd.MarkFlagRequired("output-file")
	decryptCmd.Flags().SortFlags = false
	c.AddCommand(decryptCmd)

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add encrypted secrets to a configuration properties file.",
		RunE:  c.add,
		Args:  cobra.NoArgs,
	}
	addCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	_ = addCmd.MarkFlagRequired("config-file")

	addCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	_ = addCmd.MarkFlagRequired("local-secrets-file")

	addCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	_ = addCmd.MarkFlagRequired("remote-secrets-file")

	addCmd.Flags().String("config", "", "List of configuration properties.")
	_ = addCmd.MarkFlagRequired("config")
	addCmd.Flags().SortFlags = false
	c.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update encrypted secrets from the configuration properties file.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	_ = updateCmd.MarkFlagRequired("config-file")

	updateCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	_ = updateCmd.MarkFlagRequired("local-secrets-file")

	updateCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	_ = updateCmd.MarkFlagRequired("remote-secrets-file")

	updateCmd.Flags().String("config", "", "List of configuration properties.")
	_ = updateCmd.MarkFlagRequired("config")
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete configuration values from the configuration properties file.",
		RunE:  c.remove,
		Args:  cobra.NoArgs,
	}
	removeCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	_ = removeCmd.MarkFlagRequired("config-file")

	removeCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	_ = removeCmd.MarkFlagRequired("local-secrets-file")

	removeCmd.Flags().String("config", "", "List of configuration properties.")
	_ = removeCmd.MarkFlagRequired("config")
	removeCmd.Flags().SortFlags = false
	c.AddCommand(removeCmd)

	rotateMasterKeyCmd := &cobra.Command{
		Use:   "re-encrypt",
		Short: "Rotate the master key.",
		RunE:  c.rotateMasterKey,
		Args:  cobra.NoArgs,
	}

	rotateMasterKeyCmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file. ")
	_ = rotateMasterKeyCmd.MarkFlagRequired("local-secrets-file")
	rotateMasterKeyCmd.Flags().SortFlags = false
	c.AddCommand(rotateMasterKeyCmd)

	rotateDataKeyCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate data key.",
		RunE:  c.rotateDataKey,
		Args:  cobra.NoArgs,
	}

	rotateDataKeyCmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file. ")
	_ = rotateDataKeyCmd.MarkFlagRequired("local-secrets-file")
	rotateDataKeyCmd.Flags().SortFlags = false
	c.AddCommand(rotateDataKeyCmd)
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
	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	outputPath, err := cmd.Flags().GetString("output-file")
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
		return "", fmt.Errorf("Enter the configuration properties.")
	}

	if configSource == "-" {
		reader := bufio.NewReader(os.Stdin)
		configs, err := reader.ReadString('\n')
		return configs, err
	}

	if string(configSource[0]) == "@" {
		filePath := configSource[1:]
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		configs := string(data)
		return configs, err
	}

	return "", fmt.Errorf("Invalid path to the configuration properties file.")
}

func (c *secureFileCommand) add(cmd *cobra.Command, args []string) error {
	configSource, err := cmd.Flags().GetString("config")
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

func (c *secureFileCommand) update(cmd *cobra.Command, args []string) error {
	configSource, err := cmd.Flags().GetString("config")
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

	err = c.plugin.UpdateEncryptedPasswords(configPath, localSecretsPath, remoteSecretsPath, newConfigs)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) getConfigFilePath(cmd *cobra.Command) (string, string, string, error) {
	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return "", "", "", errors.HandleCommon(err, cmd)
	}

	remoteSecretsPath, err := cmd.Flags().GetString("remote-secrets-file")
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

	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.RemoveEncryptedPasswords(configPath, localSecretsPath, removeConfigs)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) rotateMasterKey(cmd *cobra.Command, args []string) error {
	inputType := args[0]

	passphrase, err := c.getMasterKeyPassphrase(inputType)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	oldPassphrase, newPassphrase, err := c.getOldPassphrase(passphrase)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	masterKey, err := c.plugin.RotateMasterKey(oldPassphrase, newPassphrase, localSecretsPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "New Master Key: "+masterKey)
	return nil
}

func (c *secureFileCommand) rotateDataKey(cmd *cobra.Command, args []string) error {
	inputType := args[0]

	passphrase, err := c.getMasterKeyPassphrase(inputType)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.RotateDataKey(passphrase, localSecretsPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *secureFileCommand) getOldPassphrase(passphrases string) (string, string, error) {
	passphrasesArr := strings.Split(passphrases, ",")
	if len(passphrasesArr) != 2 {
		return "", "", fmt.Errorf("Missing the master key passphrase.")
	}

	return passphrasesArr[0], passphrasesArr[1], nil
}

func (c *secureFileCommand) getMasterKeyPassphrase(inputType string) (string, error) {
	passphrase := ""
	if inputType == "" {
		return passphrase, fmt.Errorf("Enter the master key passphrase.")
	}

	if inputType == "-" {
		reader := bufio.NewReader(os.Stdin)
		passphrase, err := reader.ReadString('\n')
		return passphrase, err
	}

	if strings.HasPrefix(inputType, "@") {
		filePath := inputType[1:]
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return passphrase, err
		}
		passphrase = string(data)
		return passphrase, err
	}

	return "", fmt.Errorf("Invalid master key passphrase.")
}
