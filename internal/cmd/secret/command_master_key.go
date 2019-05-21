package secret

import (
	"bufio"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	secureplugin "github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/spf13/cobra"
	"io/ioutil"
	"os"
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
			Short: "Manage master key",
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
		Short: "Create a master key",
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("passphrase", "", "Master Key Passphrase")
	_ = createCmd.MarkFlagRequired("passphrase")
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
}

func (c *masterKeyCommand) getMasterKeyPassphrase(inputType string) (string, error) {
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

	pcmd.Println(cmd, "Master Key: "+masterKey)
	return nil
}
