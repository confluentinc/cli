package secret

import (
	"os"

	"github.com/confluentinc/go-printer"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	secureplugin "github.com/confluentinc/cli/internal/pkg/secret"
)

type masterKeyCommand struct {
	*cobra.Command
	resolv pcmd.FlagResolver
	plugin secureplugin.PasswordProtection
}

// NewMasterKeyCommand returns the Cobra command for managing master key.
func NewMasterKeyCommand(resolv pcmd.FlagResolver, plugin secureplugin.PasswordProtection) *cobra.Command {
	cmd := &masterKeyCommand{
		Command: &cobra.Command{
			Use:   "master-key",
			Short: "Manage the master key for Confluent Platform.",
		},
		resolv: resolv,
		plugin: plugin,
	}
	cmd.init()
	return cmd.Command
}

func (c *masterKeyCommand) init() {
	generateCmd := &cobra.Command{
		Use:   "generate",
		Short: "Generate a master key for Confluent Platform.",
		Long:  "This command generates a master key. This key is used for encryption and decryption of configuration values.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.generate),
	}
	generateCmd.Flags().String("passphrase", "", `The key passphrase. To pipe from stdin use "-", e.g. "--passphrase -";
to read from a file use "@<path-to-file>", e.g. "--passphrase @/User/bob/secret.properties".`)
	generateCmd.Flags().SortFlags = false
	generateCmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	check(generateCmd.MarkFlagRequired("local-secrets-file"))
	c.AddCommand(generateCmd)
}

func (c *masterKeyCommand) generate(cmd *cobra.Command, _ []string) error {
	passphraseSource, err := cmd.Flags().GetString("passphrase")
	if err != nil {
		return err
	}

	passphrase, err := c.resolv.ValueFrom(passphraseSource, "Master Key Passphrase: ", true)
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

	pcmd.ErrPrintln(cmd, errors.SaveTheMasterKeyMsg)
	err = printer.RenderTableOut(&struct{ MasterKey string }{MasterKey: masterKey}, []string{"MasterKey"}, map[string]string{"MasterKey": "Master Key"}, os.Stdout)
	if err != nil {
		return err
	}
	return nil
}
