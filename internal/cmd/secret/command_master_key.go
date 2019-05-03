package secret

import (
	secureplugin "github.com/confluentinc/cli/internal/pkg/secret"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/spf13/cobra"
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
			Use:   "masterKey",
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
		Args:  cobra.ExactArgs(1),
	}
	createCmd.Flags().String("key-source", "", "choose from environment-variable/confluent-home/user-defined-path")
	_ = createCmd.MarkFlagRequired("key-source")
	c.AddCommand(createCmd)

	c.AddCommand(&cobra.Command{
		Use:   "rotate-key",
		Short: "Rotate Master Key",
		RunE:  c.rotate,
		Args:  cobra.NoArgs,
	})
}

func (c *masterKeyCommand) create(cmd *cobra.Command, args []string) error {
	keySource, err := cmd.Flags().GetString("key-source")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}


	passphrase, err := c.promptPassphrase(cmd, "Master Key Passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	path := ""
	if keySource == "user-defined-path" {
		path, err = c.promptUserPath(cmd)
		if err != nil {
			return errors.HandleCommon(err, cmd)
		}
	}

	err = c.plugin.GenerateEncryptionKeys(passphrase, keySource, path)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *masterKeyCommand) rotate(cmd *cobra.Command, args []string) error {

	oldPassphrase, err := c.promptPassphrase(cmd, "Old Master Key Passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	newPassphrase, err := c.promptPassphrase(cmd, "New Master Key Passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	err = c.plugin.RotateMasterKey(oldPassphrase, newPassphrase)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	return nil
}

func (c *masterKeyCommand) promptPassphrase(cmd *cobra.Command, display string) (string, error) {

	pcmd.Print(cmd, display)
	bytePassword, err := c.prompt.ReadPassword(0)
	if err != nil {
		return "", err
	}
	pcmd.Println(cmd)
	password := string(bytePassword)
	return password, nil
}


func (c *masterKeyCommand) promptUserPath(cmd *cobra.Command) (string, error) {

	pcmd.Print(cmd, "Enter Master Key Path: ")
	pathPrompt, err := c.prompt.ReadString('\n')
	if err != nil {
		return "", err
	}
	pcmd.Println(cmd)
	path := pathPrompt
	return  path, nil
}
