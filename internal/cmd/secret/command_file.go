package secret

import (
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"
	"os"
	"strings"
)

type secureFileCommand struct {
	*cobra.Command
	config *config.Config
	plugin secret.PasswordProtection
	prompt pcmd.Prompt
	resolv pcmd.FlagResolver
}

// NewFileCommand returns the Cobra command for managing encrypted file.
func NewFileCommand(config *config.Config, prompt pcmd.Prompt, resolv pcmd.FlagResolver, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &secureFileCommand{
		Command: &cobra.Command{
			Use:   "file",
			Short: "Secure secrets in a configuration properties file.",
		},
		config: config,
		plugin: plugin,
		prompt: prompt,
		resolv: resolv,
	}
	cmd.init()
	return cmd.Command
}

func (c *secureFileCommand) init() {
	encryptCmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt secrets in a configuration properties file.",
		Long: `This command encrypts the passwords in file specified in --config-file. This command returns a failure
if a master key has not already been set in the environment variable. Create master key using "master-key create" 
command and save the generated master key in environment variable.`,
		RunE: c.encrypt,
		Args: cobra.NoArgs,
	}
	encryptCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(encryptCmd.MarkFlagRequired("config-file"))

	encryptCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(encryptCmd.MarkFlagRequired("local-secrets-file"))

	encryptCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	check(encryptCmd.MarkFlagRequired("remote-secrets-file"))
	encryptCmd.Flags().String("config", "", "List of configuration keys.")
	encryptCmd.Flags().SortFlags = false
	c.AddCommand(encryptCmd)

	decryptCmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt encrypted secrets from the configuration properties file.",
		Long: `This command decrypts the passwords in file specified in --config-file. This command returns a failure
if a master key has not already been set using the "master-key create" command.`,
		RunE: c.decrypt,
		Args: cobra.NoArgs,
	}
	decryptCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(decryptCmd.MarkFlagRequired("config-file"))

	decryptCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(decryptCmd.MarkFlagRequired("local-secrets-file"))

	decryptCmd.Flags().String("output-file", "", "Output file path.")
	check(decryptCmd.MarkFlagRequired("output-file"))
	decryptCmd.Flags().SortFlags = false
	c.AddCommand(decryptCmd)

	addCmd := &cobra.Command{
		Use:   "add",
		Short: "Add encrypted secrets to a configuration properties file.",
		Long: `This command encrypts the password and adds it to the configuration file specified in --config-file. This
command returns a failure if a master key has not already been set using the "master-key create" command.`,
		RunE: c.add,
		Args: cobra.NoArgs,
	}
	addCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(addCmd.MarkFlagRequired("config-file"))

	addCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(addCmd.MarkFlagRequired("local-secrets-file"))

	addCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	check(addCmd.MarkFlagRequired("remote-secrets-file"))

	addCmd.Flags().String("config", "", "List of configuration properties.")
	check(addCmd.MarkFlagRequired("config"))
	addCmd.Flags().SortFlags = false
	c.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the encrypted secrets from the configuration properties file.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(updateCmd.MarkFlagRequired("config-file"))

	updateCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(updateCmd.MarkFlagRequired("local-secrets-file"))

	updateCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	check(updateCmd.MarkFlagRequired("remote-secrets-file"))

	updateCmd.Flags().String("config", "", "List of configuration properties.")
	check(updateCmd.MarkFlagRequired("config"))
	updateCmd.Flags().SortFlags = false
	c.AddCommand(updateCmd)

	removeCmd := &cobra.Command{
		Use:   "remove",
		Short: "Delete the configuration values from the configuration properties file.",
		RunE:  c.remove,
		Args:  cobra.NoArgs,
	}
	removeCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(removeCmd.MarkFlagRequired("config-file"))

	removeCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(removeCmd.MarkFlagRequired("local-secrets-file"))

	removeCmd.Flags().String("config", "", "List of configuration keys.")
	check(removeCmd.MarkFlagRequired("config"))
	removeCmd.Flags().SortFlags = false
	c.AddCommand(removeCmd)

	rotateKeyCmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate master key or data key.",
		Long: `Based on the flag set this command rotates either the master key or data key. --master-key: Generates a new master key and re-encrypts the  with the new master key. The new master
key is stored in an environment variable.
data-key: Generates a new data key and re-encrypts the file with the new data key.`,
		RunE: c.rotate,
		Args: cobra.NoArgs,
	}

	rotateKeyCmd.Flags().Bool("master-key", false, "Rotate Master Key.")
	rotateKeyCmd.Flags().Bool("data-key", false, "Rotate Data Key.")
	rotateKeyCmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file.")
	check(rotateKeyCmd.MarkFlagRequired("local-secrets-file"))
	rotateKeyCmd.Flags().String("passphrase", "", "Master key passphrase; use - to pipe from stdin or @file.txt to read from file.")
	rotateKeyCmd.Flags().SortFlags = false
	c.AddCommand(rotateKeyCmd)
}

func (c *secureFileCommand) encrypt(cmd *cobra.Command, args []string) error {
	configs, err := cmd.Flags().GetString("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	configPath, localSecretsPath, remoteSecretsPath, err := c.getConfigFilePath(cmd)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.EncryptConfigFileSecrets(configPath, localSecretsPath, remoteSecretsPath, configs)
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

func (c *secureFileCommand) add(cmd *cobra.Command, args []string) error {
	configSource, err := cmd.Flags().GetString("config")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	newConfigs, err := c.getConfigs(cmd, configSource, "config properties", "", false)
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

	newConfigs, err := c.getConfigs(cmd, configSource, "config properties", "", false)
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

	removeConfigs, err := c.getConfigs(cmd, configSource, "config properties", "", false)
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

func (c *secureFileCommand) rotate(cmd *cobra.Command, args []string) error {
	passphrasesSource, err := cmd.Flags().GetString("passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	rotateMEK, err := cmd.Flags().GetBool("master-key")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if rotateMEK {
		oldPassphrase := ""
		newPassphrase := ""
		if passphrasesSource == "" {
			oldPassphrase, err = c.getConfigs(cmd, passphrasesSource, "passphrase", "Old Master Key Passphrase: ", true)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
			newPassphrase, err = c.getConfigs(cmd, passphrasesSource, "passphrase", "New Master Key Passphrase: ", true)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}

		} else {
			passphrases, err := c.getConfigs(cmd, passphrasesSource, "passphrase", "", false)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}

			oldPassphrase, newPassphrase, err = c.getOldPassphrase(passphrases)
			if err != nil {
				return errors.HandleCommon(err, cmd)
			}
		}
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		masterKey, err := c.plugin.RotateMasterKey(oldPassphrase, newPassphrase, localSecretsPath)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		pcmd.Println(cmd, "Save the Master Key. It is not retrievable later.")
		err = printer.RenderTableOut(&struct{ MasterKey string }{MasterKey: masterKey}, []string{"MasterKey"}, map[string]string{"MasterKey": "Master Key"}, os.Stdout)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	} else {
		passphrase, err := c.getConfigs(cmd, passphrasesSource, "passphrase", "Master Key Passphrase: ", true)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		err = c.plugin.RotateDataKey(passphrase, localSecretsPath)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	return nil
}

func (c *secureFileCommand) getOldPassphrase(passphrases string) (string, string, error) {
	passphrasesArr := strings.Split(passphrases, ",")
	if len(passphrasesArr) != 2 {
		return "", "", fmt.Errorf("Missing the master key passphrase. Enter comma separated old and new master key passphrases")
	}

	return passphrasesArr[0], passphrasesArr[1], nil
}

func (c *secureFileCommand) getConfigs(cmd *cobra.Command, configSource string, inputType string, prompt string, secure bool) (string, error) {
	newConfigs, err := c.resolv.ValueFrom(configSource, prompt, secure)
	if err != nil {
		switch err {
		case pcmd.ErrNoValueSpecified:
			cmd.SilenceUsage = true
			return "", fmt.Errorf("Please enter " + inputType)
		case pcmd.ErrNoPipe:
			cmd.SilenceUsage = true
			return "", fmt.Errorf("Please pipe your " + inputType + " over stdin.")
		}
		return "", err
	}
	return newConfigs, nil
}

func check(err error) {
	if err != nil {
		panic(err)
	}
}
