package migrations

import (
	"fmt"
	"os"

	"github.com/confluentinc/cli/internal/pkg/config"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
)

func MigrateV2ToV3(cfgV2 *v2.Config) (*v3.Config, error) {
	baseCfgV3 := &config.BaseConfig{
		Params:   cfgV2.BaseConfig.Params,
		Filename: cfgV2.BaseConfig.Filename,
		Ver:      v3.Version,
	}
	cfgV3 := &v3.Config{
		BaseConfig:         baseCfgV3,
		DisableUpdateCheck: cfgV2.DisableUpdateCheck,
		DisableUpdates:     cfgV2.DisableUpdates,
		NoBrowser:          cfgV2.NoBrowser,
		Platforms:          cfgV2.Platforms,
		Credentials:        cfgV2.Credentials,
		Contexts:           nil,
		ContextStates:      cfgV2.ContextStates,
		CurrentContext:     cfgV2.CurrentContext,
		AnonymousId:        cfgV2.AnonymousId,
	}
	contextsV3 := make(map[string]*v3.Context)
	for ctxName, ctxV2 := range cfgV2.Contexts {
		contextsV3[ctxName] = migrateContextV2ToV3(ctxV2, cfgV3)
	}
	cfgV3.Contexts = contextsV3
	_, _ = fmt.Fprintf(os.Stderr, "Migrated config from V2 to V3.\n")
	if cfgV3.CLIName == "ccloud" {
		_, _ = fmt.Fprintf(os.Stderr, "Active Kafka setting and Kafka cluster information are removed from username credential contexts.\n")
	}
	return cfgV3, nil
}

func migrateContextV2ToV3(contextV2 *v2.Context, cfgV3 *v3.Config) *v3.Context {
	contextV3 := &v3.Context{
		Name:                   contextV2.Name,
		Platform:               contextV2.Platform,
		PlatformName:           contextV2.PlatformName,
		Credential:             contextV2.Credential,
		CredentialName:         contextV2.CredentialName,
		KafkaClusterContext:    nil,
		SchemaRegistryClusters: contextV2.SchemaRegistryClusters,
		State:                  contextV2.State,
		Logger:                 contextV2.Logger,
		Config:                 cfgV3,
	}
	kafka := contextV2.Kafka
	kafkaClusters := contextV2.KafkaClusters
	if cfgV3.CLIName == "ccloud" && contextV3.Credential.CredentialType == v2.Username {
		kafka = ""
		kafkaClusters = map[string]*v1.KafkaClusterConfig{}
		contextV3.Logger.Debugf("Removing active Kafka setting and Kafka cluster information from context %s as part of config migration from V2 to V3.\n", contextV3.Name)
	}
	contextV3.KafkaClusterContext = v3.NewKafkaClusterContext(contextV3, kafka, kafkaClusters)
	return contextV3
}
