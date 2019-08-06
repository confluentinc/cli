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
	key = strings.TrimSpace(key)
	fmt.Println("Enter your Schema Registry API Secret:")
	secret, err = prompt.ReadString('\n')
	if err != nil {
		return "", "", err
	}
	secret = strings.TrimSpace(secret)

	return key, secret, nil
}

func srContext(config *config.Config) (context.Context, error) {
	srCluster, err := config.SchemaRegistryCluster()
	if err != nil {
		return nil, err
	}
	if srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0 {
		key, secret, err := getSrCredentials()
		if err != nil {
			return nil, err
		}
		srCluster.SrCredentials = &configPkg.APIKeyPair{
			Key:    key,
			Secret: secret,
		}
		err = config.Save()
		if err != nil {
			return nil, err
		}
	}
	return context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: srCluster.SrCredentials.Key,
		Password: srCluster.SrCredentials.Secret,
	}), nil
}

func SchemaRegistryClient(ch *pcmd.ConfigHelper) (client *srsdk.APIClient, ctx context.Context, err error) {
	ctx, err = srContext(ch.Config)
	if err != nil {
		return nil, nil, err
	}

	srConfig := srsdk.NewConfiguration()
	if ch.Config.Auth == nil {
		return nil, nil, errors.Errorf("user must be authenticated to use Schema Registry")
	}
	srConfig.BasePath, err = ch.SchemaRegistryURL(ctx)
	if err != nil {
		return nil, nil, err
	}
	srConfig.UserAgent = ch.Version.UserAgent

	// Validate before returning
	client = srsdk.NewAPIClient(srConfig)
	_, _, err = client.DefaultApi.Get(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}

	return client, ctx, nil
}
