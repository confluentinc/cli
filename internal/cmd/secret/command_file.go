package secret

import (
	"fmt"
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/secret"
)

type secureFileCommand struct {
	*cobra.Command
	plugin secret.PasswordProtection
	prompt pcmd.Prompt
	resolv pcmd.FlagResolver
}

// NewFileCommand returns the Cobra command for managing encrypted file.
func NewFileCommand(prompt pcmd.Prompt, resolv pcmd.FlagResolver, plugin secret.PasswordProtection) *cobra.Command {
	cmd := &secureFileCommand{
		Command: &cobra.Command{
			Use:   "file",
			Short: "Secure secrets in a configuration properties file.",
		},
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
if a master key has not already been set in the environment variable. Create master key using "master-key generate"
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
if a master key has not already been set using the "master-key generate" command.`,
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
command returns a failure if a master key has not already been set using the "master-key generate" command.`,
		RunE: c.add,
		Args: cobra.NoArgs,
	}
	addCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(addCmd.MarkFlagRequired("config-file"))

	addCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(addCmd.MarkFlagRequired("local-secrets-file"))

	addCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	check(addCmd.MarkFlagRequired("remote-secrets-file"))

	addCmd.Flags().String("config", "", "List of key/value pairs of configuration properties.")
	check(addCmd.MarkFlagRequired("config"))
	addCmd.Flags().SortFlags = false
	c.AddCommand(addCmd)

	updateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update the encrypted secrets from the configuration properties file.",
		Long:  "This command updates the encrypted secrets from the configuration properties file.",
		RunE:  c.update,
		Args:  cobra.NoArgs,
	}
	updateCmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	check(updateCmd.MarkFlagRequired("config-file"))

	updateCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(updateCmd.MarkFlagRequired("local-secrets-file"))

	updateCmd.Flags().String("remote-secrets-file", "", "Path to the remote encrypted configuration properties file.")
	check(updateCmd.MarkFlagRequired("remote-secrets-file"))

	updateCmd.Flags().String("config", "", "List of key/value pairs of configuration properties.")
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
		Short: "Rotate master or data key.",
		Long: `This command rotates either the master or data key.
				To rotate the master key, specify the current master key passphrase flag ("--passphrase")
				followed by the new master key passphrase flag ("--passphrase-new").
				To rotate the data key, specify the current master key passphrase flag ("--passphrase").`,
		RunE: c.rotate,
		Args: cobra.NoArgs,
	}

	rotateKeyCmd.Flags().Bool("master-key", false, "Rotate the master key. Generates a new master key and re-encrypts with the new key.")
	rotateKeyCmd.Flags().Bool("data-key", false, "Rotate data key. Generates a new data key and re-encrypts the file with the new key.")
	rotateKeyCmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file.")
	check(rotateKeyCmd.MarkFlagRequired("local-secrets-file"))
	rotateKeyCmd.Flags().String("passphrase", "", `Master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)
	rotateKeyCmd.Flags().String("passphrase-new", "", `New master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)
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
	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	rotateMEK, err := cmd.Flags().GetBool("master-key")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	if rotateMEK {
		oldPassphraseSource, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		oldPassphrase, err := c.getConfigs(cmd, oldPassphraseSource, "passphrase", "Old Master Key Passphrase: ", true)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		newPassphraseSource, err := cmd.Flags().GetString("passphrase-new")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}

		newPassphrase, err := c.getConfigs(cmd, newPassphraseSource, "passphrase-new", "New Master Key Passphrase: ", true)
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
		passphraseSource, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
		passphrase, err := c.getConfigs(cmd, passphraseSource, "passphrase", "Master Key Passphrase: ", true)
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
