package secret

import (
	"github.com/spf13/cobra"
)

const masterKeyNotSetWarning = "This command fails if a master key has not been set in the environment variable `CONFLUENT_SECURITY_MASTER_KEY`. Create a master key using `confluent secret master-key generate`."

func (c *command) newFileCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "Secure secrets in a configuration properties file.",
	}

	cmd.AddCommand(c.newAddCommand())
	cmd.AddCommand(c.newDecryptCommand())
	cmd.AddCommand(c.newEncryptCommand())
	cmd.AddCommand(c.newRemoveCommand())
	cmd.AddCommand(c.newRotateCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func (c *command) addConfigFileFlag(cmd *cobra.Command) {
	cmd.Flags().String("config-file", "", `Path to the configuration properties file. File extension must be one of ".json" or ".properties" (key=value pairs).`)
}

func (c *command) getConfigFilePath(cmd *cobra.Command) (string, string, string, error) {
	configFile, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return "", "", "", err
	}

	localSecretsFile, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return "", "", "", err
	}

	remoteSecretsFile, err := cmd.Flags().GetString("remote-secrets-file")
	if err != nil {
		return "", "", "", err
	}

	return configFile, localSecretsFile, remoteSecretsFile, nil
}
