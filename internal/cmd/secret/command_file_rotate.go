package secret

import (
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type rotateOut struct {
	MasterKey string `human:"Master Key" serialized:"master_key"`
}

func (c *command) newRotateCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "Rotate master or data key.",
		Long:  "This command rotates either the master or data key. To rotate the master key, specify the current master key passphrase flag (`--passphrase`) followed by the new master key passphrase flag (`--passphrase-new`). To rotate the data key, specify the current master key passphrase flag (`--passphrase`).",
		Args:  cobra.NoArgs,
		RunE:  c.rotate,
	}

	cmd.Flags().String("local-secrets-file", "", "Path to the encrypted configuration properties file.")
	cmd.Flags().Bool("master-key", false, "Rotate the master key. Generates a new master key and re-encrypts with the new key.")
	cmd.Flags().Bool("data-key", false, "Rotate data key. Generates a new data key and re-encrypts the file with the new key.")
	cmd.Flags().String("passphrase", "", `Master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)
	cmd.Flags().String("passphrase-new", "", `New master key passphrase. You can use dash ("-") to pipe from stdin or @file.txt to read from file.`)

	_ = cmd.MarkFlagRequired("local-secrets-file")

	return cmd
}

func (c *command) rotate(cmd *cobra.Command, _ []string) error {
	localSecretsFile, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	masterKey, err := cmd.Flags().GetBool("master-key")
	if err != nil {
		return err
	}

	if masterKey {
		passphrase, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return err
		}

		oldPassphrase, err := c.getConfigs(passphrase, "passphrase", "Old Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		passphraseNew, err := cmd.Flags().GetString("passphrase-new")
		if err != nil {
			return err
		}

		newPassphrase, err := c.getConfigs(passphraseNew, "passphrase-new", "New Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		masterKey, err := c.plugin.RotateMasterKey(oldPassphrase, newPassphrase, localSecretsFile)
		if err != nil {
			return err
		}

		utils.ErrPrintln(errors.SaveTheMasterKeyMsg)
		table := output.NewTable(cmd)
		table.Add(&rotateOut{MasterKey: masterKey})
		return table.Print()
	} else {
		passphrase, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return err
		}

		configs, err := c.getConfigs(passphrase, "passphrase", "Master Key Passphrase: ", true)
		if err != nil {
			return err
		}

		return c.plugin.RotateDataKey(configs, localSecretsFile)
	}
}
