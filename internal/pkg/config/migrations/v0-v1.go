package migrations

import (
	"github.com/confluentinc/cli/internal/pkg/config"
	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
)

func MigrateV0ToV1(cfgV0 *v0.Config) (*v1.Config, error) {

	platformsV1 := make(map[string]*v1.Platform)
	credentialsV1 := make(map[string]*v1.Credential)
	contextsV1 := make(map[string]*v1.Context)

	for name, platformV0 := range cfgV0.Platforms {
		platformsV1[name] = migratePlatformV0ToV1(platformV0)
	}
	for name, credential := range cfgV0.Credentials {
		credentialsV1[name] = migrateCredentialV0ToV1(credential)
	}
	for name, contextV0 := range cfgV0.Contexts {
		contextsV1[name] = migrateContextV0ToV1(contextV0, name)
	}
	baseCfgV1 := &config.BaseConfig{
		Params:   cfgV0.BaseConfig.Params,
		Filename: cfgV0.BaseConfig.Filename,
		Ver:      &v1.Version,
	}
	cfgV1 := &v1.Config{
		BaseConfig:         baseCfgV1,
		DisableUpdateCheck: false,
		DisableUpdates:     false,
		AuthURL:            cfgV0.AuthURL,
		NoBrowser:          false,
		AuthToken:          cfgV0.AuthToken,
		Auth:               migrateAuthV0ToV1(cfgV0.Auth),
		Platforms:          platformsV1,
		Credentials:        credentialsV1,
		Contexts:           contextsV1,
		CurrentContext:     cfgV0.CurrentContext,
	}
	err := cfgV1.Validate()
	if err != nil {
		return nil, err
	}
	err = cfgV1.ResetAnonymousId()
	if err != nil {
		return nil, err
	}
	return cfgV1, nil
}

func migrateAuthV0ToV1(authV0 *v0.AuthConfig) *v1.AuthConfig {
	if authV0 == nil {
		return nil
	}
	return &v1.AuthConfig{
		User:     authV0.User,
		Account:  authV0.Account,
		Accounts: authV0.Accounts,
	}
}

func migratePlatformV0ToV1(platformV0 *v0.Platform) *v1.Platform {
	if platformV0 == nil {
		return nil
	}
	return &v1.Platform{
		Server:     platformV0.Server,
		CaCertPath: "",
	}
}

func migrateCredentialV0ToV1(credentialV0 *v0.Credential) *v1.Credential {
	if credentialV0 == nil {
		return nil
	}
	return &v1.Credential{
		Username:       credentialV0.Username,
		Password:       credentialV0.Password,
		APIKeyPair:     nil,
		CredentialType: v1.Username,
	}
}

func migrateContextV0ToV1(contextV0 *v0.Context, name string) *v1.Context {
	if contextV0 == nil {
		return nil
	}
	kafkaClustersV1 := make(map[string]*v1.KafkaClusterConfig)
	srClustersV1 := make(map[string]*v1.SchemaRegistryCluster)
	for name, cluster := range contextV0.KafkaClusters {
		kafkaClustersV1[name] = migrateKafkaClusterConfig(cluster)
	}
	for name, cluster := range contextV0.SchemaRegistryClusters {
		srClustersV1[name] = migrateSRCluster(cluster)
	}
	return &v1.Context{
		Name:                   name,
		Platform:               contextV0.Platform,
		Credential:             contextV0.Credential,
		KafkaClusters:          kafkaClustersV1,
		Kafka:                  contextV0.Kafka,
		SchemaRegistryClusters: srClustersV1,
	}
}

func migrateKafkaClusterConfig(clusterV0 *v0.KafkaClusterConfig) *v1.KafkaClusterConfig {
	if clusterV0 == nil {
		return nil
	}
	return &v1.KafkaClusterConfig{
		ID:          clusterV0.ID,
		Name:        clusterV0.Name,
		Bootstrap:   clusterV0.Bootstrap,
		APIEndpoint: clusterV0.APIEndpoint,
		APIKeys:     clusterV0.APIKeys,
		APIKey:      clusterV0.APIKey,
	}
}

func migrateSRCluster(srClusterV0 *v0.SchemaRegistryCluster) *v1.SchemaRegistryCluster {
	if srClusterV0 == nil {
		return nil
	}
	return &v1.SchemaRegistryCluster{
		SchemaRegistryEndpoint: srClusterV0.SchemaRegistryEndpoint,
		SrCredentials:          srClusterV0.SrCredentials,
	}
}
