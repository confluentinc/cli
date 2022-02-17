package secret

import (
	"context"
	"github.com/confluentinc/cli/internal/pkg/secret"
	"github.com/spf13/cobra"
	"net/http"
	"os"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

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

	_, exists := featureInfo.Features[secret.MdsFeatureCipherFlag]
	if exists == false {
		return secret.AES_CBC
	}

	return secret.AES_GCM
}
