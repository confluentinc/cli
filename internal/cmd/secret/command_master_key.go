package secret

import (
	"fmt"
	secureplugin "github.com/confluentinc/cli/internal/pkg/secret"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
	"bufio"
	"strings"
)

type masterKeyCommand struct {
	*cobra.Command
	config *config.Config
	prompt pcmd.Prompt
	plugin secureplugin.PasswordProtection
}

// NewMasterKeyCommand returns the Cobra command for managing master key.
func NewMasterKeyCommand(config *config.Config, prompt pcmd.Prompt, plugin secureplugin.PasswordProtection) *cobra.Command {
	cmd := &masterKeyCommand{
		Command: &cobra.Command{
			Use:   "master-key",
			Short: "Manage Master Key",
		},
		config: config,
		prompt: prompt,
		plugin: plugin,
	}
	cmd.init()
	return cmd.Command
}

func (c *masterKeyCommand) init() {
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Master Key",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("passphrase", "", "Master Key Passphrase")
	_ = createCmd.MarkFlagRequired("passphrase")
	c.AddCommand(createCmd)

	rotateMasterKeyCmd := &cobra.Command{
		Use:   "rotate-master-key",
		Short: "Rotate Master Key",
		RunE:  c.rotate_master_key,
		Args:  cobra.ExactArgs(1),
	}

	rotateMasterKeyCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = rotateMasterKeyCmd.MarkFlagRequired("local-secrets-file-path")
	c.AddCommand(rotateMasterKeyCmd)

	rotateDataKeyCmd := &cobra.Command{
		Use:   "rotate-data-key",
		Short: "Rotate Data Key",
		RunE:  c.rotate_data_key,
		Args:  cobra.ExactArgs(1),
	}

	rotateDataKeyCmd.Flags().String("local-secrets-file-path", "", "Local Encrypted Config Properties File Path.")
	_ = rotateDataKeyCmd.MarkFlagRequired("local-secrets-file-path")
	c.AddCommand(rotateDataKeyCmd)
}

func (c *masterKeyCommand) getMasterKeyPassphrase(inputType string) (string, error){
	passphrase := ""
	if inputType == "" {
		return passphrase, fmt.Errorf("Please enter master key passphrase.")
	}

	if inputType == "-" {
		reader := bufio.NewReader(os.Stdin)
		passphrase, err := reader.ReadString('\n')
		return passphrase, err
	}

	if strings.HasPrefix(inputType, "@") {
		filePath := inputType[1:len(inputType)]
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return passphrase, err
		}
		passphrase = string(data)
		return passphrase, err
	}

	return "", fmt.Errorf("Invalid master key passphrase.")
}

func (c *masterKeyCommand) create(cmd *cobra.Command, args []string) error {
	passphraseSource, err := cmd.Flags().GetString("passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	passphrase, err := c.getMasterKeyPassphrase(passphraseSource)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	masterKey, err := c.plugin.CreateMasterKey(passphrase)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "Master Key: " + masterKey)
	return nil
}

func (c *masterKeyCommand) getOldPassphrase(passphrases string) (string, string, error) {
	passphrasesArr := strings.Split(passphrases, ",")
	if len(passphrasesArr) !=2 {
		return "", "", fmt.Errorf("Missing old master key passphrase/ new master key passphrase.")
	}

	return passphrasesArr[0], passphrasesArr[1], nil
}

func (c *masterKeyCommand) rotate_master_key(cmd *cobra.Command, args []string) error {
	inputType := args[0]

	passphrase, err := c.getMasterKeyPassphrase(inputType)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	oldPassphrase, newPassphrase, err := c.getOldPassphrase(passphrase)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	masterKey, err := c.plugin.RotateMasterKey(oldPassphrase, newPassphrase, localSecretsPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "New Master Key: " + masterKey)
	return nil
}

func (c *masterKeyCommand) rotate_data_key(cmd *cobra.Command, args []string) error {
	inputType := args[0]

	passphrase, err := c.getMasterKeyPassphrase(inputType)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file-path")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.RotateDataKey(passphrase, localSecretsPath)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}
