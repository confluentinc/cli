package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v4/pkg/cmd"
	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/examples"
	"github.com/confluentinc/cli/v4/pkg/output"
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
				Text: "Generate a master key.",
				Code: "confluent secret master-key generate --local-secrets-file /path/to/secrets.txt --passphrase my-passphrase",
			},
		),
	}

	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("passphrase", "", "The key passphrase.")
	pcmd.AddOutputFlag(cmd)

	cobra.CheckErr(cmd.MarkFlagRequired("local-secrets-file"))
	cobra.CheckErr(cmd.MarkFlagRequired("passphrase"))

	return cmd
}

func (c *command) generate(cmd *cobra.Command, _ []string) error {
	passphrase, err := cmd.Flags().GetString("passphrase")
	if err != nil {
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
