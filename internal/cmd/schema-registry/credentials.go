package schema_registry

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/confluentinc/ccloud-sdk-go"
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

func srContext(cfg *config.Config, client *ccloud.Client) (context.Context, error) {
	srCluster, err := cfg.SchemaRegistryCluster(client)
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

func SchemaRegistryClient(cfg *config.Config, client *ccloud.Client) (srClient *srsdk.APIClient, ctx context.Context, err error) {
	ctx, err = srContext(cfg, client)
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
		ctxClient := config.NewContextClient(currCtx, client)
		srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, state.Auth.Account.Id)
		if err != nil {
			return nil, nil, err
		}
		srConfig.BasePath = srCluster.Endpoint
	}
	srConfig.UserAgent = currCtx.Version.UserAgent
	// Validate before returning.
	srClient = srsdk.NewAPIClient(srConfig)
	_, _, err = srClient.DefaultApi.Get(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}
	return srClient, ctx, nil
}
