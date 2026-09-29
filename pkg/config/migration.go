package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// legacyAPIKeyPair holds the secret-bearing fields of a v4 APIKeyPair, retagged json:"-" at HEAD.
type legacyAPIKeyPair struct {
	Key    string `json:"api_key,omitempty"`
	Secret string `json:"api_secret,omitempty"`
	Salt   []byte `json:"salt,omitempty"`
	Nonce  []byte `json:"nonce,omitempty"`
}

// applyTo copies o's secret fields onto live as-is (ciphertext stays ciphertext).
func (o *legacyAPIKeyPair) applyTo(live *APIKeyPair) {
	if o == nil || live == nil {
		return
	}
	live.Secret = o.Secret
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyContextState holds the secret-bearing fields of a v4 ContextState, retagged json:"-" at
// HEAD.
type legacyContextState struct {
	AuthToken        string `json:"auth_token"`
	AuthRefreshToken string `json:"auth_refresh_token"`
	Salt             []byte `json:"salt,omitempty"`
	Nonce            []byte `json:"nonce,omitempty"`
}

func (o *legacyContextState) applyTo(live *ContextState) {
	if o == nil || live == nil {
		return
	}
	live.AuthToken = o.AuthToken
	live.AuthRefreshToken = o.AuthRefreshToken
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyLoginCredential holds the secret-bearing fields of a v4 LoginCredential, retagged
// json:"-" at HEAD.
type legacyLoginCredential struct {
	EncryptedPassword string `json:"encrypted_password"`
	Salt              []byte `json:"salt,omitempty"`
	Nonce             []byte `json:"nonce,omitempty"`
}

func (o *legacyLoginCredential) applyTo(live *LoginCredential) {
	if o == nil || live == nil {
		return
	}
	live.EncryptedPassword = o.EncryptedPassword
	live.Salt = o.Salt
	live.Nonce = o.Nonce
}

// legacyCredential reaches Credential.APIKeyPair, the only secret-bearing field under Credential.
// The deprecated Credential.Password field (v4 "password,omitempty") is not decoded: nothing
// reads it at HEAD, and it was never wired into the live Credential type.
type legacyCredential struct {
	APIKeyPair *legacyAPIKeyPair `json:"api_key_pair"`
}

// legacyKafkaClusterConfig reaches the API keys nested under a Kafka cluster config, keyed by
// API key, whether the config lives under kafka_cluster_configs or kafka_environment_contexts.
type legacyKafkaClusterConfig struct {
	APIKeys map[string]*legacyAPIKeyPair `json:"api_keys"`
}

// applyKafkaClusterConfigs copies secrets from legacy Kafka cluster configs onto their live
// counterparts, matched by cluster id and then by API key.
func applyKafkaClusterConfigs(legacyConfigs map[string]*legacyKafkaClusterConfig, live map[string]*KafkaClusterConfig) {
	for id, lcfg := range legacyConfigs {
		if lcfg == nil {
			continue
		}
		cfg, ok := live[id]
		if !ok {
			continue
		}
		for key, pair := range lcfg.APIKeys {
			if livePair, ok := cfg.APIKeys[key]; ok {
				pair.applyTo(livePair)
			}
		}
	}
}

// legacyKafkaEnvContext reaches Kafka cluster configs nested under an environment, whose v4 tag
// (kafka_cluster_infos) differs from the top-level kafka_cluster_configs tag.
type legacyKafkaEnvContext struct {
	KafkaClusterConfigs map[string]*legacyKafkaClusterConfig `json:"kafka_cluster_infos"`
}

// legacyKafkaClusterContext reaches both places Kafka cluster configs (and their nested API
// keys) can live: directly, and per environment.
type legacyKafkaClusterContext struct {
	KafkaClusterConfigs map[string]*legacyKafkaClusterConfig `json:"kafka_cluster_configs,omitempty"`
	KafkaEnvContexts    map[string]*legacyKafkaEnvContext    `json:"kafka_environment_contexts,omitempty"`
}

// legacySchemaRegistryCluster reaches the deprecated per-context Schema Registry credential.
type legacySchemaRegistryCluster struct {
	SrCredentials *legacyAPIKeyPair `json:"schema_registry_credentials"`
}

// legacyContext reaches every secret- or cache-bearing field nested under a v4 Context.
// FeatureFlags decodes directly into the live *FeatureFlags type: none of its own fields are
// retagged, only its holder (Context.FeatureFlags) is.
type legacyContext struct {
	FeatureFlags           *FeatureFlags                           `json:"feature_flags,omitempty"`
	GlobalAPIKeys          map[string]*legacyAPIKeyPair            `json:"global_api_keys,omitempty"`
	KafkaClusterContext    *legacyKafkaClusterContext              `json:"kafka_cluster_context"`
	SchemaRegistryClusters map[string]*legacySchemaRegistryCluster `json:"schema_registry_clusters,omitempty"`
}

// legacyConfig is the overlay for a v4 config.json: every secret- and cache-bearing field that
// HEAD retagged json:"-" or moved off Config. Decoding a v4 file into this type, alongside a
// normal decode into a live Config, recovers exactly the values HEAD's live decode drops.
type legacyConfig struct {
	LastUpdateCheckAt *time.Time                        `json:"last_update_check_at,omitempty"`
	Credentials       map[string]*legacyCredential      `json:"credentials,omitempty"`
	Contexts          map[string]*legacyContext         `json:"contexts,omitempty"`
	ContextStates     map[string]*legacyContextState    `json:"context_states,omitempty"`
	SavedCredentials  map[string]*legacyLoginCredential `json:"saved_credentials,omitempty"`
}

// applyTo copies every overlay value onto c's live, already-decoded structures, matched by the
// same map keys the JSON used. A key present in the overlay but absent from c (shouldn't happen,
// since both are decoded from the same bytes) is silently skipped rather than fabricating one.
func (o *legacyConfig) applyTo(c *Config) {
	c.LastUpdateCheckAt = o.LastUpdateCheckAt

	for name, lcred := range o.Credentials {
		if lcred == nil {
			continue
		}
		if cred, ok := c.Credentials[name]; ok {
			lcred.APIKeyPair.applyTo(cred.APIKeyPair)
		}
	}

	for name, lctx := range o.Contexts {
		ctx, ok := c.Contexts[name]
		if !ok || lctx == nil {
			continue
		}

		ctx.FeatureFlags = lctx.FeatureFlags

		for key, pair := range lctx.GlobalAPIKeys {
			if live, ok := ctx.GlobalAPIKeys[key]; ok {
				pair.applyTo(live)
			}
		}

		if lctx.KafkaClusterContext != nil && ctx.KafkaClusterContext != nil {
			applyKafkaClusterConfigs(lctx.KafkaClusterContext.KafkaClusterConfigs, ctx.KafkaClusterContext.KafkaClusterConfigs)

			for envName, lenv := range lctx.KafkaClusterContext.KafkaEnvContexts {
				if lenv == nil {
					continue
				}
				if env, ok := ctx.KafkaClusterContext.KafkaEnvContexts[envName]; ok {
					applyKafkaClusterConfigs(lenv.KafkaClusterConfigs, env.KafkaClusterConfigs)
				}
			}
		}

		for id, lsr := range lctx.SchemaRegistryClusters {
			if lsr == nil {
				continue
			}
			if sr, ok := ctx.SchemaRegistryClusters[id]; ok {
				lsr.SrCredentials.applyTo(sr.SrCredentials)
			}
		}
	}

	for name, lstate := range o.ContextStates {
		if lstate == nil {
			continue
		}
		if state, ok := c.ContextStates[name]; ok {
			lstate.applyTo(state)
		}
	}

	for name, lcred := range o.SavedCredentials {
		if lcred == nil {
			continue
		}
		if cred, ok := c.SavedCredentials[name]; ok {
			lcred.applyTo(cred)
		}
	}
}

// readLegacyConfigFile reads a v4 config.json at path. It returns (nil, false, nil) when the
// file is missing, not a regular file (following symlinks), or zero bytes; any other read
// failure is a hard error.
func readLegacyConfigFile(path string) ([]byte, bool, error) {
	info, err := os.Stat(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, false, nil
		}
		return nil, false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	if !info.Mode().IsRegular() || info.Size() == 0 {
		return nil, false, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false, fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}
	if len(data) == 0 {
		return nil, false, nil
	}
	return data, true, nil
}

// applyLegacyConfig decodes a v4 config.json's bytes onto c, secrets included. It parses the
// live shape and the secret overlay into scratch values first, so a parse error leaves c
// untouched; only once both succeed does it decode onto c directly (preserving New()'s defaults
// for keys absent from the file, and c.Filename, exactly like v4's own decode-onto-c) and copy
// the overlay's values into c's json:"-" fields. It never wires contexts or saves.
func (c *Config) applyLegacyConfig(path string, data []byte) error {
	live := New()
	if err := json.Unmarshal(data, live); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	overlay := new(legacyConfig)
	if err := json.Unmarshal(data, overlay); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	if err := json.Unmarshal(data, c); err != nil {
		return fmt.Errorf(errors.UnableToReadConfigurationFileErrorMsg, path, err)
	}

	overlay.applyTo(c)
	return nil
}
