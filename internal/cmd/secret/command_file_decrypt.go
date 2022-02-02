package secret

import (
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
)

func (c *command) newDecryptCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt secrets in a configuration properties file.",
		Long:  "This command decrypts the passwords in file specified in `--config-file`. This command returns a failure if a master key has not already been set using the `master-key generate` command.",
		Args:  cobra.NoArgs,
		RunE:  pcmd.NewCLIRunE(c.decrypt),
	}

	cmd.Flags().String("config-file", "", "Path to the configuration properties file.")
	cmd.Flags().String("local-secrets-file", "", "Path to the local encrypted configuration properties file.")
	cmd.Flags().String("output-file", "", "Output file path.")
	cmd.Flags().String("config", "", "List of configuration keys.")

	_ = cmd.MarkFlagRequired("config-file")
	_ = cmd.MarkFlagRequired("local-secrets-file")
	_ = cmd.MarkFlagRequired("output-file")

	return cmd
}

func (c *command) decrypt(cmd *cobra.Command, _ []string) error {
	configs, err := cmd.Flags().GetString("config")
	if err != nil {
		return err
	}

	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return err
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return err
	}

	outputPath, err := cmd.Flags().GetString("output-file")
	if err != nil {
		return err
	}

	return c.plugin.DecryptConfigFileSecrets(configPath, localSecretsPath, outputPath, configs)
}
