package schema_registry

import (
	"context"
	"fmt"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	configPkg "github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"os"
	"strings"
)

func getSrCredentials() (key string, secret string, err error) {
	prompt := pcmd.NewPrompt(os.Stdin)
	fmt.Println("Enter your Schema Registry API Key:")
	key, err = prompt.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	fmt.Println("Enter your Schema Registry API Secret:")
	secret, err = prompt.ReadString('\n')
	if err != nil {
		return "", "", err
	}

	// Validate before returning
	_, _, err = srsdk.APIClient{}.DefaultApi.Get(context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: key,
		Password: secret,
	}))
	if err != nil {
		return "", "", errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}
	return strings.TrimSpace(key), strings.TrimSpace(secret), nil
}

func SrContext(config *config.Config) (context.Context, error) {
	if config.SrCredentials == nil || len(config.SrCredentials.Key) == 0 || len(config.SrCredentials.Secret) == 0 {
		key, secret, err := getSrCredentials()
		if err != nil {
			return nil, err
		}
		config.SrCredentials = &configPkg.APIKeyPair{
			Key:    key,
			Secret: secret,
		}
		config.Save()
	}
	return context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: config.SrCredentials.Key,
		Password: config.SrCredentials.Secret,
	}), nil
}

func SchemaRegistryClient(ch *pcmd.ConfigHelper) (client *srsdk.APIClient, err error) {
	srConfig := srsdk.NewConfiguration()
	if ch.Config.Auth == nil {
		return nil, errors.Errorf("user must be authenticated to use Schema Registry")
	}
	srConfig.BasePath, err = ch.SchemaRegistryURL(ch.Config.Auth.Account.Id)
	if err != nil {
		return nil, err
	}
	//srConfig.UserAgent = version.UserAgent
	return srsdk.NewAPIClient(srConfig), nil
}
