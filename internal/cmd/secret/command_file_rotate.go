package secret

import (
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newRotateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate master or data key.",
		Long:  "This command rotates either the master or data key. To rotate the master key, specify the current master key passphrase flag (`--passphrase`) followed by the new master key passphrase flag (`--passphrase-new`). To rotate the data key, specify the current master key passphrase flag (`--passphrase`).",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.rotate),
	}

	cmd.Flags().Bool("master-key", false, "Rotate the master key. Generates a new master key and re-encrypts with the new key.")
	cmd.Flags().Bool("data-key", false, "Rotate data key. Generates a new data key and re-encrypts the file with the new key.")
	cmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file.")
	cmd.Flags().String("passphrase", "", `Master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)
	cmd.Flags().String("passphrase-new", "", `New master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)

	_ = cmd.MarkFlagRequired("local-secrets-file")

	return cmd
}

func (c *command) rotate(cmd *cobra.Command, _ []string) error {
	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	rotateMEK, err := cmd.Flags().GetBool("master-key")
	if err != nil {
		return err
	}

	cipherMode := c.getCipherMode()
	c.plugin.SetCipherMode(cipherMode)

	if rotateMEK {
		oldPassphraseSource, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return err
		}

		oldPassphrase, err := c.getConfigs(oldPassphraseSource, "passphrase", "Old Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		newPassphraseSource, err := cmd.Flags().GetString("passphrase-new")
		if err != nil {
			return err
		}

		newPassphrase, err := c.getConfigs(newPassphraseSource, "passphrase-new", "New Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		masterKey, err := c.plugin.RotateMasterKey(oldPassphrase, newPassphrase, localSecretsPath)
		if err != nil {
			return err
		}

		utils.ErrPrintln(cmd, "Save the Master Key. It is not retrievable later.")
		return printer.RenderTableOut(&struct{ MasterKey string }{MasterKey: masterKey}, []string{"MasterKey"}, map[string]string{"MasterKey": "Master Key"}, os.Stdout)
	} else {
		passphraseSource, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return err
		}

		passphrase, err := c.getConfigs(passphraseSource, "passphrase", "Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		return c.plugin.RotateDataKey(passphrase, localSecretsPath)
	}
}
