package secret

import (
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func (c *command) newGenerateFunction() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a master key for Confluent Platform.",
		Long:  "This command generates a master key. This key is used for encryption and decryption of configuration values.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.generate),
	}

	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("passphrase", "", `The key passphrase. To pipe from stdin use "-", e.g. "--passphrase -". To read from a file use "@<path-to-file>", e.g. "--passphrase @/User/bob/secret.properties".`)

	_ = cmd.MarkFlagRequired("local-secrets-file")

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
			return errors.New(errors.SpecifyPassphraseErrorMsg)
		case pcmd.ErrNoPipe:
			return errors.New(errors.PipePassphraseErrorMsg)
		}
		return err
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	masterKey, err := c.plugin.CreateMasterKey(passphrase, localSecretsPath)
	if err != nil {
		return err
	}

	utils.ErrPrintln(cmd, errors.SaveTheMasterKeyMsg)
	return printer.RenderTableOut(&struct{ MasterKey string }{MasterKey: masterKey}, []string{"MasterKey"}, map[string]string{"MasterKey": "Master Key"}, os.Stdout)
}
