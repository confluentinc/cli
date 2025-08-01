package config

import (
	"fmt"

	"github.com/confluentinc/cli/v4/pkg/output"
)

type KafkaClusterContext struct {
	EnvContext bool `json:"environment_context"`
	// ActiveKafkaCluster is your active Kafka cluster and references a key in the KafkaClusters map
	ActiveKafkaCluster string `json:"active_kafka,omitempty"`
	// ActiveKafkaClusterEndpoint is your active endpoint corresponding to your active Kafka cluster
	ActiveKafkaClusterEndpoint string `json:"active_kafka_endpoint,omitempty"`
	// KafkaClusterConfigs store connection info for interacting directly with Kafka (e.g., consume/produce, etc)
	// N.B. These may later be exposed in the CLI to directly register kafkas (outside a Control Plane)
	// Mapped by cluster id.
	KafkaClusterConfigs map[string]*KafkaClusterConfig `json:"kafka_cluster_configs,omitempty"`
	KafkaEnvContexts    map[string]*KafkaEnvContext    `json:"kafka_environment_contexts,omitempty"`
	Context             *Context                       `json:"-"`
}

type KafkaEnvContext struct {
	ActiveKafkaCluster         string                         `json:"active_kafka"`
	ActiveKafkaClusterEndpoint string                         `json:"active_kafka_endpoint"`
	KafkaClusterConfigs        map[string]*KafkaClusterConfig `json:"kafka_cluster_infos"`
}

func NewKafkaClusterContext(ctx *Context, activeKafka string, activeKafkaEndpoint string, kafkaClusters map[string]*KafkaClusterConfig) *KafkaClusterContext {
	if ctx.IsCloud(ctx.Config.IsTest) && ctx.GetCredentialType() == Username {
		return newKafkaClusterEnvironmentContext(activeKafka, activeKafkaEndpoint, kafkaClusters, ctx)
	} else {
		return newKafkaClusterNonEnvironmentContext(activeKafka, activeKafkaEndpoint, kafkaClusters, ctx)
	}
}

func newKafkaClusterEnvironmentContext(activeKafka string, activeKafkaEndpoint string, kafkaClusters map[string]*KafkaClusterConfig, ctx *Context) *KafkaClusterContext {
	return &KafkaClusterContext{
		EnvContext: true,
		KafkaEnvContexts: map[string]*KafkaEnvContext{ctx.GetCurrentEnvironment(): {
			ActiveKafkaCluster:         activeKafka,
			ActiveKafkaClusterEndpoint: activeKafkaEndpoint,
			KafkaClusterConfigs:        kafkaClusters,
		}},
		Context: ctx,
	}
}

func newKafkaClusterNonEnvironmentContext(activeKafka string, activeKafkaEndpoint string, kafkaClusters map[string]*KafkaClusterConfig, ctx *Context) *KafkaClusterContext {
	return &KafkaClusterContext{
		EnvContext:                 false,
		ActiveKafkaCluster:         activeKafka,
		ActiveKafkaClusterEndpoint: activeKafkaEndpoint,
		KafkaClusterConfigs:        kafkaClusters,
		Context:                    ctx,
	}
}

func (k *KafkaClusterContext) GetActiveKafkaClusterId() string {
	if !k.EnvContext {
		return k.ActiveKafkaCluster
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.ActiveKafkaCluster
}

func (k *KafkaClusterContext) GetActiveKafkaClusterConfig() *KafkaClusterConfig {
	if !k.EnvContext {
		return k.KafkaClusterConfigs[k.ActiveKafkaCluster]
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.KafkaClusterConfigs[kafkaEnvContext.ActiveKafkaCluster]
}

func (k *KafkaClusterContext) GetActiveKafkaClusterEndpoint() string {
	if !k.EnvContext {
		return k.ActiveKafkaClusterEndpoint
	}
	kafkaEnvContext := k.GetCurrentKafkaEnvContext()
	return kafkaEnvContext.ActiveKafkaClusterEndpoint
}

func (k *KafkaClusterContext) SetActiveKafkaCluster(clusterId string) {
	if !k.EnvContext {
		k.ActiveKafkaCluster = clusterId
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		kafkaEnvContext.ActiveKafkaCluster = clusterId
	}
}

func (k *KafkaClusterContext) SetActiveKafkaClusterEndpoint(endpoint string) {
	if !k.EnvContext {
		k.ActiveKafkaClusterEndpoint = endpoint
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		kafkaEnvContext.ActiveKafkaClusterEndpoint = endpoint
	}
}

func (k *KafkaClusterContext) UnsetActiveKafkaClusterEndpoint() {
	if !k.EnvContext {
		k.ActiveKafkaClusterEndpoint = ""
	} else {
		kafkaEnvContext := k.GetCurrentKafkaEnvContext()
		kafkaEnvContext.ActiveKafkaClusterEndpoint = ""
	}
}

func (k *KafkaClusterContext) GetKafkaClusterConfig(clusterId string) *KafkaClusterConfig {
	if !k.EnvContext {
		return k.KafkaClusterConfigs[clusterId]
	}

	return k.GetCurrentKafkaEnvContext().KafkaClusterConfigs[clusterId]
}

func (k *KafkaClusterContext) AddKafkaClusterConfig(kcc *KafkaClusterConfig) {
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

func (k *KafkaClusterContext) FindApiKeyClusterId(key string) string {
	clusterConfigs := k.KafkaClusterConfigs
	if k.EnvContext {
		clusterConfigs = k.GetCurrentKafkaEnvContext().KafkaClusterConfigs
	}

	for id, config := range clusterConfigs {
		for apiKey := range config.APIKeys {
			if key == apiKey {
				return id
			}
		}
	}

	return ""
}

func (k *KafkaClusterContext) DeleteApiKey(key string) {
	clusterConfigs := k.KafkaClusterConfigs
	if k.EnvContext {
		clusterConfigs = k.GetCurrentKafkaEnvContext().KafkaClusterConfigs
	}

	if id := k.FindApiKeyClusterId(key); id != "" {
		delete(clusterConfigs[id].APIKeys, key)
		if clusterConfigs[id].APIKey == key {
			clusterConfigs[id].APIKey = ""
		}
	}
}

func (k *KafkaClusterContext) GetCurrentKafkaEnvContext() *KafkaEnvContext {
	curEnv := k.Context.GetCurrentEnvironment()
	if k.KafkaEnvContexts[curEnv] == nil {
		k.KafkaEnvContexts[curEnv] = &KafkaEnvContext{
			ActiveKafkaCluster:         "",
			ActiveKafkaClusterEndpoint: "",
			KafkaClusterConfigs:        map[string]*KafkaClusterConfig{},
		}
		if err := k.Context.Save(); err != nil {
			panic(fmt.Sprintf("Unable to save new KafkaEnvContext to config for context '%s' environment '%s'.", k.Context.Name, curEnv))
		}
	}
	return k.KafkaEnvContexts[curEnv]
}

func (k *KafkaClusterContext) Validate() {
	k.validateActiveKafka()
	if !k.EnvContext {
		if k.KafkaClusterConfigs == nil {
			k.KafkaClusterConfigs = map[string]*KafkaClusterConfig{}
			if err := k.Context.Save(); err != nil {
				panic(fmt.Sprintf("Unable to save new KafkaClusterConfigs map to config for context '%s'.", k.Context.Name))
			}
		}
		for _, kcc := range k.KafkaClusterConfigs {
			k.validateKafkaClusterConfig(kcc)
		}
	} else {
		if k.KafkaEnvContexts == nil {
			k.KafkaEnvContexts = map[string]*KafkaEnvContext{}
			if err := k.Context.Save(); err != nil {
				panic(fmt.Sprintf("Unable to save new KafkaEnvContexts map to config for context '%s'.", k.Context.Name))
			}
		}
		for env, kafkaEnvContexts := range k.KafkaEnvContexts {
			if kafkaEnvContexts.KafkaClusterConfigs == nil {
				kafkaEnvContexts.KafkaClusterConfigs = map[string]*KafkaClusterConfig{}
				if err := k.Context.Save(); err != nil {
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
	errMsg := "Active Kafka cluster \"%s\" has no info stored for context \"%s\".\n" +
		"Removing active Kafka setting for the context.\n" +
		"You can set the active Kafka cluster with `confluent kafka cluster use`.\n"
	if !k.EnvContext {
		if _, ok := k.KafkaClusterConfigs[k.ActiveKafkaCluster]; k.ActiveKafkaCluster != "" && !ok {
			output.ErrPrintf(false, errMsg, k.ActiveKafkaCluster, k.Context.Name)
			k.ActiveKafkaCluster = ""
			if err := k.Context.Save(); err != nil {
				panic(fmt.Sprintf("Unable to reset ActiveKafkaCluster in context '%s'.", k.Context.Name))
			}
		}
	} else {
		for env, kafkaEnvContext := range k.KafkaEnvContexts {
			if _, ok := kafkaEnvContext.KafkaClusterConfigs[kafkaEnvContext.ActiveKafkaCluster]; kafkaEnvContext.ActiveKafkaCluster != "" && !ok {
				output.ErrPrintf(false, errMsg, kafkaEnvContext.ActiveKafkaCluster, k.Context.Name)
				kafkaEnvContext.ActiveKafkaCluster = ""
				if err := k.Context.Save(); err != nil {
					panic(fmt.Sprintf("Unable to reset ActiveKafkaCluster in context '%s', environment '%s'.", k.Context.Name, env))
				}
			}
		}
	}
}

func (k *KafkaClusterContext) validateKafkaClusterConfig(cluster *KafkaClusterConfig) {
	if cluster.ID == "" {
		panic(fmt.Sprintf("cluster under context '%s' has no id", k.Context.Name))
	}
	if cluster.APIKeys == nil {
		cluster.APIKeys = map[string]*APIKeyPair{}
		if err := k.Context.Save(); err != nil {
			panic(fmt.Sprintf("Unable to save new APIKeys map in context '%s', for cluster '%s'.", k.Context.Name, cluster.ID))
		}
	}
	if _, ok := cluster.APIKeys[cluster.APIKey]; cluster.APIKey != "" && !ok {
		output.ErrPrintf(false, "Current API key \"%s\" of resource \"%s\" under context \"%s\" is not found.\n", cluster.APIKey, cluster.ID, k.Context.Name)
		output.ErrPrintln(false, "Removing current API key setting for the resource.")
		output.ErrPrintf(false, "You can re-add the API key with `confluent api-key store --resource %s` and then set current API key with `confluent api-key use`.\n", cluster.ID)
		cluster.APIKey = ""
		if err := k.Context.Save(); err != nil {
			panic(fmt.Sprintf("Unable to reset current APIKey for cluster '%s' in context '%s'.", cluster.ID, k.Context.Name))
		}
	}
	k.validateApiKeysDict(cluster)
}

func (k *KafkaClusterContext) validateApiKeysDict(cluster *KafkaClusterConfig) {
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
		if err := k.Context.Save(); err != nil {
			panic("Unable to save new KafkaEnvContext to config.")
		}
	}
}
