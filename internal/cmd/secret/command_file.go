package secret

import (
	"context"
	"net/http"
	"os"

	"github.com/spf13/cobra"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/secret"
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

func (c *command) getConfigFilePath(cmd *cobra.Command) (string, string, string, error) {
	configPath, err := cmd.Flags().GetString("config-file")
	if err != nil {
		return "", "", "", err
	}

	localSecretsPath, err := cmd.Flags().GetString("local-secrets-file")
	if err != nil {
		return "", "", "", err
	}

	remoteSecretsPath, err := cmd.Flags().GetString("remote-secrets-file")
	if err != nil {
		return "", "", "", err
	}

	return configPath, localSecretsPath, remoteSecretsPath, nil
}

func (c *command) getConfigs(configSource string, inputType string, prompt string, secure bool) (string, error) {
	newConfigs, err := c.flagResolver.ValueFrom(configSource, prompt, secure)
	if err != nil {
		switch err {
		case pcmd.ErrNoValueSpecified:
			return "", errors.Errorf(errors.EnterInputTypeErrorMsg, inputType)
		case pcmd.ErrNoPipe:
			return "", errors.Errorf(errors.PipeInputTypeErrorMsg, inputType)
		}
		return "", err
	}
	return newConfigs, nil
}

func (c *command) getCipherMode() string {
	if os.Getenv("XX_SECRETS_GCM_MODE") != "" {
		return secret.AES_GCM
	}

	ctx := context.WithValue(context.Background(), mds.ContextAccessToken, c.State.AuthToken)
	featureInfo, response, err := c.MDSClient.MetadataServiceOperationsApi.Features(ctx)

	if err != nil || response.StatusCode == http.StatusNotFound {
		return secret.AES_CBC
	}

	if _, ok := featureInfo.Features[secret.MdsFeatureCipherFlag]; ok {
		return secret.AES_GCM
	}
	return secret.AES_CBC
}
