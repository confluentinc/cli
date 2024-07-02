package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
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
	cmd.Flags().String("passphrase", "", `Master key passphrase.`)
	cmd.Flags().String("passphrase-new", "", `New master key passphrase.`)
	cmd.Flags().Bool("master-key", false, "Rotate the master key. Generates a new master key and re-encrypts with the new key.")
	cmd.Flags().Bool("data-key", false, "Rotate data key. Generates a new data key and re-encrypts the file with the new key.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("passphrase"))
	cobra.CheckErr(cmd.MarkFlagRequired("passphrase-new"))

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

		passphraseNew, err := cmd.Flags().GetString("passphrase-new")
		if err != nil {
			return err
		}

		masterKey, err := c.plugin.RotateMasterKey(passphrase, passphraseNew, localSecretsFile)
		if err != nil {
			return err
		}

		output.ErrPrintln(c.Config.EnableColor, errors.SaveTheMasterKeyMsg)
		table := output.NewTable(cmd)
		table.Add(&rotateOut{MasterKey: masterKey})
		return table.Print()
	} else {
		passphrase, err := cmd.Flags().GetString("passphrase")
		if err != nil {
			return err
		}

		return c.plugin.RotateDataKey(passphrase, localSecretsFile)
	}
}
