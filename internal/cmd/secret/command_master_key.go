package secret

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	secureplugin "github.com/confluentinc/cli/internal/pkg/secret"
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
			Short: "Manage the master key for Confluent Platform.",
			Long: "Manage the master keys with this command.",
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
		Short: "Create a master key for Confluent Platform.",
		Long: `This command generates a master key and stores it in an environment variable. This key will be used for encryption and
decryption of configuration value.`,
		RunE:  c.create,
		Args:  cobra.NoArgs,
	}
	createCmd.Flags().String("passphrase", "", `The key passphrase. To pipe from stdin use "-", e.g. "--passphrase -";
to read from a file use "@<path-to-file>", e.g. "--passphrase @/User/bob/secret.properties".`)
	createCmd.Flags().SortFlags = false
	c.AddCommand(createCmd)
}

func (c *masterKeyCommand) getMasterKeyPassphrase(cmd *cobra.Command, source string) (string, error) {
	if source == "" {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			// TODO: should we require this or just assume that pipe to stdin implies '--passphrase -' ?
			return "", fmt.Errorf("To pipe your passphrase over stdin, specify '--passphrase -'.")
		} else {
			pcmd.Print(cmd, "Master Key Passphrase: ")
			passphrase, err := c.prompt.ReadPassword()
			pcmd.Println(cmd, "\n")
			return passphrase, err
		}
	}

	if source == "-" {
		fi, _ := os.Stdin.Stat()
		if (fi.Mode() & os.ModeCharDevice) == 0 {
			reader := bufio.NewReader(os.Stdin)
			passphrase, err := reader.ReadString('\n')
			return passphrase, err
		} else {
			return "", fmt.Errorf("You must pipe your passphrase over stdin, e.g. \"--passphrase -\".")
		}
	}

	if strings.HasPrefix(source, "@") {
		filePath := source[1:]
		data, err := ioutil.ReadFile(filePath)
		if err != nil {
			return "", err
		}
		passphrase := string(data)
		return passphrase, err
	}

	return source, nil
}

func (c *masterKeyCommand) create(cmd *cobra.Command, args []string) error {
	passphraseSource, err := cmd.Flags().GetString("passphrase")
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	passphrase, err := c.getMasterKeyPassphrase(cmd, passphraseSource)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	masterKey, err := c.plugin.CreateMasterKey(passphrase)
	if err != nil {
		return errors.HandleCommon(err, cmd)
	}

	pcmd.Println(cmd, "Save the master key. It cannot be retrievable later.")
	printer.RenderTableOut(&struct{MasterKey string}{MasterKey: masterKey}, []string{"MasterKey"}, map[string]string{"MasterKey": "Master Key"}, os.Stdout)
	return nil
}
