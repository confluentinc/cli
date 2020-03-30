package migrations

import (
	"strings"

	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
)

func MigrateV1ToV2(cfgV1 *v1.Config) (*v2.Config, error) {
	platformsV1 := make(map[string]*v2.Platform)
	for name, platformV0 := range cfgV1.Platforms {
		platformsV1[name] = migratePlatformV1ToV2(platformV0)
	}
	credentialsV1 := make(map[string]*v2.Credential)
	for name, credentialV0 := range cfgV1.Credentials {
		credentialsV1[name] = migrateCredentialV1ToV2(credentialV0)
	}
	baseCfgV2 := &config.BaseConfig{
		Params:   cfgV1.BaseConfig.Params,
		Filename: cfgV1.BaseConfig.Filename,
		Ver:      v2.Version,
	}
	cfgV2 := &v2.Config{
		BaseConfig:         baseCfgV2,
		DisableUpdateCheck: cfgV1.DisableUpdateCheck,
		DisableUpdates:     cfgV1.DisableUpdates,
		NoBrowser:          cfgV1.DisableUpdates,
		Platforms:          platformsV1,
		Credentials:        credentialsV1,
		Contexts:           nil,
		ContextStates:      nil,
		CurrentContext:     cfgV1.CurrentContext,
		AnonymousId:        cfgV1.AnonymousId,
	}
	contextsV1 := make(map[string]*v2.Context)
	contextStates := make(map[string]*v2.ContextState)
	for name, contextV0 := range cfgV1.Contexts {
		contextV1, state := migrateContextV1ToV2(contextV0, platformsV1[contextV0.Platform], credentialsV1[contextV0.Credential], cfgV1, cfgV2)
		contextsV1[name] = contextV1
		contextStates[name] = state
	}
	cfgV2.Contexts = contextsV1
	cfgV2.ContextStates = contextStates
	err := cfgV2.Validate()
	if err != nil {
		return nil, err
	}
	return cfgV2, nil
}

func migrateContextV1ToV2(contextV1 *v1.Context, platformV2 *v2.Platform, credentialV2 *v2.Credential, cfgV1 *v1.Config, cfgV2 *v2.Config) (*v2.Context, *v2.ContextState) {
	if contextV1 == nil {
		return nil, nil
	}
	srClustersV1 := make(map[string]*v2.SchemaRegistryCluster)
	for envId, srClusterV0 := range contextV1.SchemaRegistryClusters {
		srClustersV1[envId] = migrateSRClusterV1ToV2(srClusterV0)
	}
	state := &v2.ContextState{
		Auth:      cfgV1.Auth,
		AuthToken: cfgV1.AuthToken,
	}
	contextV2 := &v2.Context{
		Name:                   contextV1.Name,
		Platform:               platformV2,
		PlatformName:           contextV1.Platform,
		Credential:             credentialV2,
		CredentialName:         contextV1.Credential,
		KafkaClusters:          contextV1.KafkaClusters,
		Kafka:                  contextV1.Kafka,
		SchemaRegistryClusters: srClustersV1,
		State:                  state,
		Logger:                 cfgV1.Logger,
		Config:                 cfgV2,
	}
	return contextV2, state
}

func migrateSRClusterV1ToV2(srClusterV1 *v1.SchemaRegistryCluster) *v2.SchemaRegistryCluster {
	if srClusterV1 == nil {
		return nil
	}
	srClusterV2 := &v2.SchemaRegistryCluster{
		Id:                     "",
		SchemaRegistryEndpoint: srClusterV1.SchemaRegistryEndpoint,
		SrCredentials:          srClusterV1.SrCredentials,
	}
	return srClusterV2
}

func migratePlatformV1ToV2(platformV1 *v1.Platform) *v2.Platform {
	if platformV1 == nil {
		return nil
	}
	platformV2 := &v2.Platform{
		Name:       strings.TrimPrefix(platformV1.Server, "https://"),
		Server:     platformV1.Server,
		CaCertPath: platformV1.CaCertPath,
	}
	return platformV2
}

func migrateCredentialV1ToV2(credentialV1 *v1.Credential) *v2.Credential {
	if credentialV1 == nil {
		return nil
	}
	credentialV2 := &v2.Credential{
		Name:           credentialV1.String(),
		Username:       credentialV1.Username,
		Password:       credentialV1.Password,
		APIKeyPair:     credentialV1.APIKeyPair,
		CredentialType: migrateCredentialTypeV1ToV2(credentialV1.CredentialType),
	}
	return credentialV2
}

func migrateCredentialTypeV1ToV2(credTypeV1 v1.CredentialType) v2.CredentialType {
	switch credTypeV1 {
	case v1.Username:
		return v2.Username
	case v1.APIKey:
		return v2.APIKey
	default:
		panic("unknown credential type")
	}
}
