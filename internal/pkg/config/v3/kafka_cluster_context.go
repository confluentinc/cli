package v3

import (
	"fmt"
	"os"

	v0 "github.com/confluentinc/cli/internal/pkg/config/v0"
	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	v2 "github.com/confluentinc/cli/internal/pkg/config/v2"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

type KafkaClusterContext struct {
	EnvContext bool `json:"environment_context"`
	// ActiveKafkaCluster is your active Kafka cluster and references a key in the KafkaClusters map
	ActiveKafkaCluster string `json:"active_kafka,omitempty"`
	// KafkaClusterConfigs store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	// Mapped by cluster id.
	KafkaClusterConfigs map[string]*v1.KafkaClusterConfig `json:"kafka_cluster_configs,omitempty"`
	KafkaEnvContexts    map[string]*KafkaEnvContext       `json:"kafka_environment_contexts,omitempty"`
	Context             *Context                          `json:"-"`
}

type KafkaEnvContext struct {
	ActiveKafkaCluster  string                            `json:"active_kafka"`
	KafkaClusterConfigs map[string]*v1.KafkaClusterConfig `json:"kafka_cluster_infos"`
}

func NewKafkaClusterContext(ctx *Context, activeKafka string, kafkaClusters map[string]*v1.KafkaClusterConfig) *KafkaClusterContext {
	if ctx.Config.CLIName == "ccloud" && ctx.Credential.CredentialType == v2.Username {
		return newKafkaClusterEnvironmentContext(activeKafka, kafkaClusters, ctx)
	} else {
		return newKafkaClusterNonEnvironmentContext(activeKafka, kafkaClusters, ctx)
	}

}

func newKafkaClusterEnvironmentContext(activeKafka string, kafkaClusters map[string]*v1.KafkaClusterConfig, ctx *Context) *KafkaClusterContext {
	kafkaEnvContext := &KafkaEnvContext{
		ActiveKafkaCluster:  activeKafka,
		KafkaClusterConfigs: kafkaClusters,
	}
	kafkaClusterContext := &KafkaClusterContext{
		EnvContext:       true,
		KafkaEnvContexts: map[string]*KafkaEnvContext{ctx.GetCurrentEnvironmentId(): kafkaEnvContext},
		Context:          ctx,
	}
	return kafkaClusterContext
}

func newKafkaClusterNonEnvironmentContext(activeKafka string, kafkaClusters map[string]*v1.KafkaClusterConfig, ctx *Context) *KafkaClusterContext {
	kafkaClusterContext := &KafkaClusterContext{
		EnvContext:          false,
		ActiveKafkaCluster:  activeKafka,
		KafkaClusterConfigs: kafkaClusters,
		Context:             ctx,
	}
	return kafkaClusterContext
}

func (k *KafkaClusterContext) GetActiveKafkaClusterId() string {
	if !k.EnvContext {
		return k.ActiveKafkaCluster
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.ActiveKafkaCluster
}

func (k *KafkaClusterContext) GetActiveKafkaClusterConfig() *v1.KafkaClusterConfig {
	if !k.EnvContext {
		return k.KafkaClusterConfigs[k.ActiveKafkaCluster]
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.KafkaClusterConfigs[kafkaEnvContext.ActiveKafkaCluster]
}

func (k *KafkaClusterContext) SetActiveKafkaCluster(clusterId string) {
	if !k.EnvContext {
		k.ActiveKafkaCluster = clusterId
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		kafkaEnvContext.ActiveKafkaCluster = clusterId
	}
}

func (k *KafkaClusterContext) GetKafkaClusterConfig(clusterId string) *v1.KafkaClusterConfig {
	if !k.EnvContext {
		return k.KafkaClusterConfigs[clusterId]
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.KafkaClusterConfigs[clusterId]
}

func (k *KafkaClusterContext) AddKafkaClusterConfig(kcc *v1.KafkaClusterConfig) {
	if !k.EnvContext {
		k.KafkaClusterConfigs[kcc.ID] = kcc
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		kafkaEnvContext.KafkaClusterConfigs[kcc.ID] = kcc
	}
}

func (k *KafkaClusterContext) RemoveKafkaCluster(clusterId string) {
	if !k.EnvContext {
		delete(k.KafkaClusterConfigs, clusterId)
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		delete(kafkaEnvContext.KafkaClusterConfigs, clusterId)
	}
	if clusterId == k.GetActiveKafkaClusterId() {
		k.SetActiveKafkaCluster("")
	}
}

func (k *KafkaClusterContext) DeleteAPIKey(apiKey string) {
	var clusterConfigs map[string]*v1.KafkaClusterConfig
	if !k.EnvContext {
		clusterConfigs = k.KafkaClusterConfigs
	} else {
		clusterConfigs = k.GetCurrentKafkaEnvContext().KafkaClusterConfigs
	}
	for _, kcc := range clusterConfigs {
		for clusterApiKey := range kcc.APIKeys {
			if apiKey == clusterApiKey {
				delete(kcc.APIKeys, apiKey)
			}
			if apiKey == kcc.APIKey {
				kcc.APIKey = ""
			}
		}
	}
}

func (k *KafkaClusterContext) GetCurrentKafkaEnvContext() *KafkaEnvContext {
	curEnv := k.Context.GetCurrentEnvironmentId()
	if k.KafkaEnvContexts[curEnv] == nil {
		k.KafkaEnvContexts[curEnv] = &KafkaEnvContext{
			ActiveKafkaCluster:  "",
			KafkaClusterConfigs: map[string]*v1.KafkaClusterConfig{},
		}
		err := k.Context.Save()
		if err != nil {
			panic(fmt.Sprintf("Unable to save new KafkaEnvContext to config for context '%s' environment '%s'.", k.Context.Name, curEnv))
		}
	}
	return k.KafkaEnvContexts[curEnv]
}

func (k *KafkaClusterContext) Validate() {
	k.validateActiveKafka()
	if !k.EnvContext {
		if k.KafkaClusterConfigs == nil {
			k.KafkaClusterConfigs = map[string]*v1.KafkaClusterConfig{}
			err := k.Context.Save()
			if err != nil {
				panic(fmt.Sprintf("Unable to save new KafkaClusterConfigs map to config for context '%s'.", k.Context.Name))
			}
		}
		for _, kcc := range k.KafkaClusterConfigs {
			k.validateKafkaClusterConfig(kcc)
		}
	} else {
		if k.KafkaEnvContexts == nil {
			k.KafkaEnvContexts = map[string]*KafkaEnvContext{}
			err := k.Context.Save()
			if err != nil {
				panic(fmt.Sprintf("Unable to save new KafkaEnvContexts map to config for context '%s'.", k.Context.Name))
			}
		}
		for env, kafkaEnvContexts := range k.KafkaEnvContexts {
			if kafkaEnvContexts.KafkaClusterConfigs == nil {
				kafkaEnvContexts.KafkaClusterConfigs = map[string]*v1.KafkaClusterConfig{}
				err := k.Context.Save()
				if err != nil {
					panic(fmt.Sprintf("Unable to save new KafkaClusterConfigs map to config for context '%s', environment '%s'.", k.Context.Name, env))
				}
			}
			for _, kcc := range kafkaEnvContexts.KafkaClusterConfigs {
				k.validateKafkaClusterConfig(kcc)
			}
		}

	}
}

func (k *KafkaClusterContext) validateActiveKafka() {
	errMsg := "Active Kafka cluster '%s' has no info stored in config for context '%s'.\n" +
		"Removing active Kafka setting for the context.\n" +
		"You can set active Kafka cluster with 'ccloud kafka cluster use'.\n"
	if !k.EnvContext {
		if _, ok := k.KafkaClusterConfigs[k.ActiveKafkaCluster]; k.ActiveKafkaCluster != "" && !ok {
			_, _ = fmt.Fprintf(os.Stderr, errMsg, k.ActiveKafkaCluster, k.Context.Name)
			k.ActiveKafkaCluster = ""
			err := k.Context.Save()
			if err != nil {
				panic(fmt.Sprintf("Unable to reset ActiveKafkaCluster in context '%s'.", k.Context.Name))
			}
		}
	} else {
		for env, kafkaEnvContext := range k.KafkaEnvContexts {
			if _, ok := kafkaEnvContext.KafkaClusterConfigs[kafkaEnvContext.ActiveKafkaCluster]; kafkaEnvContext.ActiveKafkaCluster != "" && !ok {
				_, _ = fmt.Fprintf(os.Stderr, errMsg, kafkaEnvContext.ActiveKafkaCluster, k.Context.Name)
				k.ActiveKafkaCluster = ""
				err := k.Context.Save()
				if err != nil {
					panic(fmt.Sprintf("Unable to reset ActiveKafkaCluster in context '%s', environment '%s'.", k.Context.Name, env))
				}
			}
		}
	}
}

func (k *KafkaClusterContext) validateKafkaClusterConfig(cluster *v1.KafkaClusterConfig) {
	if cluster.ID == "" {
		panic(fmt.Sprintf("cluster under context '%s' has no id", k.Context.Name))
	}
	if cluster.APIKeys == nil {
		cluster.APIKeys = map[string]*v0.APIKeyPair{}
		err := k.Context.Save()
		if err != nil {
			panic(fmt.Sprintf("Unable to save new APIKeys map in context '%s', for cluster '%s'.", k.Context.Name, cluster.ID))
		}
	}
	if _, ok := cluster.APIKeys[cluster.APIKey]; cluster.APIKey != "" && !ok {
		_, _ = fmt.Fprintf(os.Stderr, errors.CurrentAPIKeyAutofixMsg, cluster.APIKey, cluster.ID, k.Context.Name, cluster.ID)
		cluster.APIKey = ""
		err := k.Context.Save()
		if err != nil {
			panic(fmt.Sprintf("Unable to reset current APIKey for cluster '%s' in context '%s'.", cluster.ID, k.Context.Name))
		}
	}
	k.validateApiKeysDict(cluster)
}

func (k *KafkaClusterContext) validateApiKeysDict(cluster *v1.KafkaClusterConfig) {
	missingKey := false
	mismatchKey := false
	missingSecret := false
	for k, pair := range cluster.APIKeys {
		if pair.Key == "" {
			delete(cluster.APIKeys, k)
			missingKey = true
			continue
		}
		if k != pair.Key {
			delete(cluster.APIKeys, k)
			mismatchKey = true
			continue
		}
		if pair.Secret == "" {
			delete(cluster.APIKeys, k)
			missingSecret = true
		}
	}
	if missingKey || mismatchKey || missingSecret {
		printApiKeysDictErrorMessage(missingKey, mismatchKey, missingSecret, cluster, k.Context.Name)
		err := k.Context.Save()
		if err != nil {
			panic("Unable to save new KafkaEnvContext to config.")
		}
	}
}
