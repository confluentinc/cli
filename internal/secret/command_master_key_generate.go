package secret

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newGenerateFunction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a master key for Confluent Platform.",
		Long:  "This command generates a master key. This key is used for encryption and decryption of configuration values.",
		Args:  cobra.NoArgs,
		RunE:  c.generate,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Pipe the passphrase from stdin:`,
				Code: "confluent secret master-key generate --local-secrets-file /path/to/secrets.txt --passphrase -",
			},
			examples.Example{
				Text: `Read the passphrase from the file "/User/bob/secret.properties":`,
				Code: "confluent secret master-key generate --local-secrets-file /path/to/secrets.txt --passphrase @/User/bob/secret.properties",
			},
		),
	}

	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("passphrase", "", "The key passphrase.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))

	return cmd
}

func (c *command) generate(cmd *cobra.Command, _ []string) error {
	passphraseSource, err := cmd.Flags().GetString("passphrase")
	if err != nil {
		return err
	}

	passphrase, err := c.flagResolver.ValueFrom(passphraseSource, "Master Key Passphrase: ", true)
	if err != nil {
		switch err {
		case pcmd.ErrUnexpectedStdinPipe:
			// TODO: should we require this or just assume that pipe to stdin implies '--passphrase -' ?
			return fmt.Errorf("specify `--passphrase -` if you intend to pipe your passphrase over stdin")
		case pcmd.ErrNoPipe:
			return fmt.Errorf("pipe your passphrase over stdin")
		}
		return err
	}

	localSecretsFile, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	masterKey, err := c.plugin.CreateMasterKey(passphrase, localSecretsFile)
	if err != nil {
		return err
	}

	output.ErrPrintln(c.Config.EnableColor, errors.SaveTheMasterKeyMsg)
	table := output.NewTable(cmd)
	table.Add(&rotateOut{MasterKey: masterKey})
	return table.Print()
}
