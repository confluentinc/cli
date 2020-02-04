package migrations

import (
	"strings"

	"github.com/confluentinc/cli/internal/pkg/config/v0"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
)

func MigrateV0ToV1(cfgV0 *v0.Config) (*v1.Config, error) {
	platformsV1 := make(map[string]*v1.Platform)
	for name, platformV0 := range cfgV0.Platforms {
		platformsV1[name] = migratePlatform(platformV0)
	}
	credentialsV1 := make(map[string]*v1.Credential)
	for name, credentialV0 := range cfgV0.Credentials {
		credentialsV1[name] = migrateCredential(credentialV0)
	}
	cfgV1 := &v1.Config{
		Params:             cfgV0.Params,
		Filename:           cfgV0.Filename,
		DisableUpdateCheck: cfgV0.DisableUpdateCheck,
		DisableUpdates:     cfgV0.DisableUpdates,
		NoBrowser:          cfgV0.DisableUpdates,
		Platforms:          platformsV1,
		Credentials:        credentialsV1,
		Contexts:           nil,
		ContextStates:      nil,
		CurrentContext:     cfgV0.CurrentContext,
		AnonymousId:        cfgV0.AnonymousId,
	}
	contextsV1 := make(map[string]*v1.Context)
	contextStates := make(map[string]*v1.ContextState)
	for name, contextV0 := range cfgV0.Contexts {
		contextV1, state := migrateContext(contextV0, platformsV1[contextV0.Platform], credentialsV1[contextV0.Credential], cfgV0, cfgV1)
		contextsV1[name] = contextV1
		contextStates[name] = state
	}
	cfgV1.Contexts = contextsV1
	cfgV1.ContextStates = contextStates
	err := cfgV1.Save()
	if err != nil {
		return nil, err
	}
	return cfgV1, nil
}

func migrateContext(contextV0 *v0.Context, platformV1 *v1.Platform, credentialV1 *v1.Credential, cfgV0 *v0.Config, cfgV1 *v1.Config) (*v1.Context, *v1.ContextState) {
	srClustersV1 := make(map[string]*v1.SchemaRegistryCluster)
	for envId, srClusterV0 := range contextV0.SchemaRegistryClusters {
		srClustersV1[envId] = migrateSRCluster(srClusterV0)
	}
	state := &v1.ContextState{
		Auth:      cfgV0.Auth,
		AuthToken: cfgV0.AuthToken,
	}
	contextV1 := &v1.Context{
		Name:                   contextV0.Name,
		Platform:               platformV1,
		PlatformName:           contextV0.Platform,
		Credential:             credentialV1,
		CredentialName:         contextV0.Credential,
		KafkaClusters:          contextV0.KafkaClusters,
		Kafka:                  contextV0.Kafka,
		SchemaRegistryClusters: srClustersV1,
		State:                  state,
		Logger:                 cfgV0.Logger,
		Config:                 cfgV1,
	}
	return contextV1, state
}

func migrateSRCluster(srClusterV0 *v0.SchemaRegistryCluster) *v1.SchemaRegistryCluster {
	srClusterV1 := &v1.SchemaRegistryCluster{
		Id:                     "",
		SchemaRegistryEndpoint: srClusterV0.SchemaRegistryEndpoint,
		SrCredentials:          srClusterV0.SrCredentials,
	}
	return srClusterV1
}

func migratePlatform(platformV0 *v0.Platform) *v1.Platform {
	platformV1 := &v1.Platform{
		Name:       strings.TrimPrefix(platformV0.Server, "https://"),
		Server:     platformV0.Server,
		CaCertPath: platformV0.CaCertPath,
	}
	return platformV1
}

func migrateCredential(credentialV0 *v0.Credential) *v1.Credential {
	credentialV1 := &v1.Credential{
		Name:           credentialV0.String(),
		Username:       credentialV0.Username,
		Password:       credentialV0.Password,
		APIKeyPair:     credentialV0.APIKeyPair,
		CredentialType: migrateCredentialType(credentialV0.CredentialType),
	}
	return credentialV1
}

func migrateCredentialType(credTypeV0 v0.CredentialType) v1.CredentialType {
	switch credTypeV0 {
	case v0.Username:
		return v1.Username
	case v0.APIKey:
		return v1.APIKey
	default:
		panic("unknown credential type")
	}
}
