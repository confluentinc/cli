package schema_registry

import (
	"context"
	"fmt"
	"os"
	"strings"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/version"
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

func srContext(cfg *pcmd.DynamicConfig, cmd *cobra.Command) (context.Context, error) {
	ctx, err := cfg.Context(cmd)
	if err != nil {
		return nil, err
	}
	srCluster, err := ctx.SchemaRegistryCluster(cmd)
	if err != nil {
		return nil, err
	}
	if srCluster.SrCredentials == nil || len(srCluster.SrCredentials.Key) == 0 || len(srCluster.SrCredentials.Secret) == 0 {
		key, secret, err := getSrCredentials()
		if err != nil {
			return nil, err
		}
		srCluster.SrCredentials = &v0.APIKeyPair{
			Key:    key,
			Secret: secret,
		}
		err = ctx.Save()
		if err != nil {
			return nil, err
		}
	}
	return context.WithValue(context.Background(), srsdk.ContextBasicAuth, srsdk.BasicAuth{
		UserName: srCluster.SrCredentials.Key,
		Password: srCluster.SrCredentials.Secret,
	}), nil
}

func SchemaRegistryClient(cmd *cobra.Command, cfg *pcmd.DynamicConfig, ver *version.Version) (srClient *srsdk.APIClient, ctx context.Context, err error) {
	ctx, err = srContext(cfg, cmd)
	if err != nil {
		return nil, nil, err
	}
	srConfig := srsdk.NewConfiguration()
	currCtx, err := cfg.Context(cmd)
	if err != nil {
		return nil, nil, err
	}
	envId, err := currCtx.AuthenticatedEnvId(cmd)
	if err != nil {
		return nil, nil, err
	}
	if srCluster, ok := currCtx.SchemaRegistryClusters[envId]; ok {
		srConfig.BasePath = srCluster.SchemaRegistryEndpoint
	} else {
		ctxClient := pcmd.NewContextClient(currCtx)
		srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(ctx, envId)
		if err != nil {
			return nil, nil, err
		}
		srConfig.BasePath = srCluster.Endpoint
	}
	srConfig.UserAgent = ver.UserAgent
	// validate before returning.
	srClient = srsdk.NewAPIClient(srConfig)
	_, _, err = srClient.DefaultApi.Get(ctx)
	if err != nil {
		return nil, nil, errors.Errorf("Failed to validate Schema Registry API Key and Secret")
	}
	return srClient, ctx, nil
}
