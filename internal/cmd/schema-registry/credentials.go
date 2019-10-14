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

func SchemaRegistryClient(cfg *config.Config) (client *srsdk.APIClient, ctx context.Context, err error) {
	ctx, err = srContext(cfg)
	if err != nil {
		return nil, nil, err
	}
	srConfig := srsdk.NewConfiguration()
	state, err := cfg.AuthenticatedState()
	if err != nil {
		return nil, nil, err
	}
	currCtx := cfg.Context()
	if srCluster, ok := currCtx.SchemaRegistryClusters[state.Auth.Account.Id]; ok {
		srConfig.BasePath = srCluster.SchemaRegistryEndpoint
	} else {
		ctxClient := config.NewContextClient(currCtx, nil)
		srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(state.Auth.Account.Id, ctx)
		if err != nil {
			return nil, nil, err
		}
		srConfig.BasePath = srCluster.Endpoint
	}
	srConfig.UserAgent = currCtx.Version.UserAgent
	// Validate before returning
	client = srsdk.NewAPIClient(srConfig)
	_, _, err = client.DefaultApi.Get(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}
	return client, ctx, nil
}
