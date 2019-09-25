package schema_registry

import (
	"context"
	"fmt"
	"os"
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/errors"
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

func srContext(cfg *config.Config) (context.Context, error) {
	srCluster, err := cfg.SchemaRegistryCluster()
	if err != nil {
		return nil, err
	}
	if srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0 {
		key, secret, err := getSrCredentials()
		if err != nil {
			return nil, err
		}
		srCluster.SrCredentials = &config.APIKeyPair{
			Key:    key,
			Secret: secret,
		}
		err = cfg.Save()
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
	err = ch.Config.CheckLogin()
	if err != nil {
		return nil, nil, err
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
