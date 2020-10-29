package schema_registry

import (
	"context"
	"os"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/form"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"github.com/confluentinc/cli/internal/pkg/version"
)

func promptSchemaRegistryCredentials(command *cobra.Command) (string, string, error) {
	f := form.New(
		form.Field{ID: "api-key", Prompt: "Enter your Schema Registry API key"},
		form.Field{ID: "secret", Prompt: "Enter your Schema Registry API secret", IsHidden: true},
	)
	if err := f.Prompt(command, form.NewPrompt(os.Stdin)); err != nil {
		return "", "", err
	}
	return f.Responses["api-key"].(string), f.Responses["secret"].(string), nil
}

func getSchemaRegistryAuth(cmd *cobra.Command, srCredentials *v0.APIKeyPair, shouldPrompt bool) (*srsdk.BasicAuth, bool, error) {
	auth := &srsdk.BasicAuth{}
	didPromptUser := false

	if srCredentials != nil {
		auth.UserName = srCredentials.Key
		auth.Password = srCredentials.Secret
	}

	if auth.UserName == "" || auth.Password == "" || shouldPrompt {
		var err error
		auth.UserName, auth.Password, err = promptSchemaRegistryCredentials(cmd)
		if err != nil {
			return nil, false, err
		}
		didPromptUser = true
	}

	return auth, didPromptUser, nil
}

func getSchemaRegistryClient(cmd *cobra.Command, cfg *pcmd.DynamicConfig, ver *version.Version) (*srsdk.APIClient, context.Context, error) {
	srConfig := srsdk.NewConfiguration()

	currCtx, err := cfg.Context(cmd)
	if err != nil {
		return nil, nil, err
	}

	srCluster := &v2.SchemaRegistryCluster{}
	endpoint, _ := cmd.Flags().GetString("sr-endpoint")
	if len(endpoint) == 0 {
		srCluster, err = currCtx.SchemaRegistryCluster(cmd)
		if err != nil {
			return nil, nil, err
		}
	}

	// First examine existing credentials. If check fails(saved credentials no longer works or user enters
	//incorrect information), shouldPrompt becomes true and prompt users to enter credentials again.
	shouldPrompt := false

	for {
		// Get credentials as Schema Registry BasicAuth
		srAuth, didPromptUser, err := getSchemaRegistryAuth(cmd, srCluster.SrCredentials, shouldPrompt)
		if err != nil {
			return nil, nil, err
		}
		srCtx := context.WithValue(context.Background(), srsdk.ContextBasicAuth, *srAuth)

		if len(endpoint) == 0 {
			envId, err := currCtx.AuthenticatedEnvId(cmd)
			if err != nil {
				return nil, nil, err
			}

			if srCluster, ok := currCtx.SchemaRegistryClusters[envId]; ok {
				srConfig.BasePath = srCluster.SchemaRegistryEndpoint
			} else {
				ctxClient := pcmd.NewContextClient(currCtx)
				srCluster, err := ctxClient.FetchSchemaRegistryByAccountId(srCtx, envId)
				if err != nil {
					return nil, nil, err
				}
				srConfig.BasePath = srCluster.Endpoint
			}
		} else {
			srConfig.BasePath = endpoint
		}
		srConfig.UserAgent = ver.UserAgent
		srClient := srsdk.NewAPIClient(srConfig)

		// Test credentials
		if _, _, err = srClient.DefaultApi.Get(srCtx); err != nil {
			utils.ErrPrintln(cmd, errors.SRCredsValidationFailedMsg)
			// Prompt users to enter new credentials if validation fails.
			shouldPrompt = true
			continue
		}

		if didPromptUser {
			// Save credentials
			srCluster.SrCredentials = &v0.APIKeyPair{
				Key:    srAuth.UserName,
				Secret: srAuth.Password,
			}
			if err := currCtx.Save(); err != nil {
				return nil, nil, err
			}
		}

		return srClient, srCtx, nil
	}
}
