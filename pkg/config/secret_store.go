package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// secretRecord holds one credential identity's secret material as stored on disk: each
// value is ciphertext paired with its own salt/nonce. A credential identity can be shared
// by several contexts (a saved password, GlobalAPIKeys, and nested Kafka API keys are all
// identity-scoped), so each secret kind carries its own salt/nonce pair rather than
// sharing one - pairing a ciphertext with the wrong salt/nonce fails GCM authentication on
// load. Auth tokens are NOT identity-scoped: they are endpoint-scoped and keyed by context
// name instead, in tokenRecord/secretFile.Tokens - see tokenRecord's doc comment.
type secretRecord struct {
	// Secret is the api-key secret, from APIKeyPair.
	Secret      string `json:"secret,omitempty"`
	SecretSalt  []byte `json:"secret_salt,omitempty"`
	SecretNonce []byte `json:"secret_nonce,omitempty"`

	// Password is the saved login password, from LoginCredential.
	Password      string `json:"password,omitempty"`
	PasswordSalt  []byte `json:"password_salt,omitempty"`
	PasswordNonce []byte `json:"password_nonce,omitempty"`

	// GlobalAPIKeys holds org-scoped API keys (Context.GlobalAPIKeys), keyed by API key id.
	GlobalAPIKeys map[string]*apiKeySecret `json:"global_api_keys,omitempty"`

	// KafkaAPIKeys holds per-cluster API keys (KafkaClusterConfig.APIKeys), keyed by cluster id
	// then by API key id.
	KafkaAPIKeys map[string]map[string]*apiKeySecret `json:"kafka_api_keys,omitempty"`
}

// tokenRecord holds one context's auth tokens as stored on disk, keyed by CONTEXT NAME
// (secretFile.Tokens), not credential identity. A token is encrypted with the context
// name as GCM associated data and is endpoint-scoped, while the credential identity
// (secretRecord's key) is username-only: two contexts can share one identity (same
// username, different URL/CA-cert) while holding two different tokens. Keying tokens by
// identity would broadcast one context's token onto the other on load - GCM
// authentication then fails on Unix (wrong AAD), and DPAPI on Windows ignores AAD
// entirely and would silently load the wrong token.
type tokenRecord struct {
	AuthToken        string `json:"auth_token,omitempty"`
	AuthRefreshToken string `json:"auth_refresh_token,omitempty"`
	Salt             []byte `json:"salt,omitempty"`
	Nonce            []byte `json:"nonce,omitempty"`
}

// apiKeySecret is a nested API key's secret material, as stored on disk: ciphertext paired
// with its own salt/nonce.
type apiKeySecret struct {
	Secret string `json:"secret,omitempty"`
	Salt   []byte `json:"salt,omitempty"`
	Nonce  []byte `json:"nonce,omitempty"`
}

// secretFile is the on-disk shape of the secret store. Secrets is keyed by credential
// identity (Context.identityKey) so contexts sharing a login share one entry and a context
// rename never orphans it. Tokens is keyed by context name - see tokenRecord.
type secretFile struct {
	Secrets map[string]*secretRecord `json:"secrets,omitempty"`
	Tokens  map[string]*tokenRecord  `json:"tokens,omitempty"`
}

// secretStore reads and writes the encrypted secret file at path.
type secretStore struct{ path string }

// newSecretStore builds a store for the default secrets file location.
func newSecretStore() *secretStore {
	return &secretStore{path: SecretsFilename()}
}

// write persists file to the store's path, creating its parent directory first.
func (s *secretStore) write(file *secretFile) error {
	dir := filepath.Dir(s.path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return fmt.Errorf("unable to create secret store directory %s: %w", dir, err)
	}

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("unable to marshal secret store: %w", err)
	}

	return writeFileAtomic(s.path, data)
}

// identityKey is the rename-stable key for a context's secrets: CredentialName is set once at
// context creation and untouched by a later context rename.
func (c *Context) identityKey() string {
	return c.CredentialName
}

// allKafkaClusterConfigs returns every KafkaClusterConfig reachable from k, keyed by cluster
// id, across both shapes KafkaClusterContext.Validate walks: a single non-env map, or one map
// per environment for a cloud username/password context.
func allKafkaClusterConfigs(k *KafkaClusterContext) map[string]*KafkaClusterConfig {
	if k == nil {
		return nil
	}
	if !k.EnvContext {
		return k.KafkaClusterConfigs
	}
	all := map[string]*KafkaClusterConfig{}
	for _, envContext := range k.KafkaEnvContexts {
		for id, kcc := range envContext.KafkaClusterConfigs {
			all[id] = kcc
		}
	}
	return all
}

// encryptedAPIKeySecret returns pair's secret/salt/nonce as an apiKeySecret triple, encrypting
// a local copy first if pair's own secret is still plaintext (EncryptSecret is a no-op if it
// is already ciphertext). nil, nil means pair has no secret to store. The live pair itself is
// never mutated.
func encryptedAPIKeySecret(pair *APIKeyPair) (*apiKeySecret, error) {
	if pair == nil || pair.Secret == "" {
		return nil, nil
	}
	shadow := &APIKeyPair{Key: pair.Key, Secret: pair.Secret, Salt: pair.Salt, Nonce: pair.Nonce}
	if err := shadow.EncryptSecret(); err != nil {
		return nil, err
	}
	return &apiKeySecret{Secret: shadow.Secret, Salt: shadow.Salt, Nonce: shadow.Nonce}, nil
}

// rehydrateNestedAPIKeySecretPresence guards merged's nested API-key structural fields
// (GlobalAPIKeys, KafkaClusterConfig.APIKeys) against Validate()'s "missing secret" cleanup.
// merged is rebuilt through threeWayMerge's JSON-based deep copies, which drop every json:"-"
// field including these secrets, so merged's own copy of a real, live, encrypted key always
// reads as empty - Validate() would then delete the key's public id along with its "missing"
// secret, even though the real encrypted secret is correctly on its way to the secret store
// via saveSecretStore(c). This sets presence only (any non-empty value; merged's own Secret
// field is json:"-" and never reaches the write below either way) for a key this process's own
// live c already knows about. A key only a concurrent session added exists only on disk and
// is not rehydrated - reconciling that is task 2.4's three-way merge, not this.
func rehydrateNestedAPIKeySecretPresence(c, merged *Config) {
	for name, ctx := range c.Contexts {
		mergedCtx, ok := merged.Contexts[name]
		if !ok || mergedCtx == nil {
			continue
		}

		for keyId, pair := range ctx.GlobalAPIKeys {
			if pair == nil || pair.Secret == "" {
				continue
			}
			if mergedPair, ok := mergedCtx.GlobalAPIKeys[keyId]; ok && mergedPair != nil {
				mergedPair.Secret = pair.Secret
			}
		}

		mergedClusters := allKafkaClusterConfigs(mergedCtx.KafkaClusterContext)
		for clusterId, cluster := range allKafkaClusterConfigs(ctx.KafkaClusterContext) {
			mergedCluster, ok := mergedClusters[clusterId]
			if !ok || mergedCluster == nil {
				continue
			}
			for keyId, pair := range cluster.APIKeys {
				if pair == nil || pair.Secret == "" {
					continue
				}
				if mergedPair, ok := mergedCluster.APIKeys[keyId]; ok && mergedPair != nil {
					mergedPair.Secret = pair.Secret
				}
			}
		}
	}
}

// saveSecretStore extracts c's secret material into the secret store, keyed by identity
// (Secrets) or by context name (Tokens). It reads the live c directly, never a
// merged/deep-copied Config: every secret field is json:"-", so json.Marshal (the copy
// mechanism threeWayMerge's inputs go through) silently drops it, and a value read back from
// that copy is always empty - encrypting that empty value would quietly replace the real
// secret with ciphertext of "".
//
// A credential's API secret and a context's auth tokens toggle to plaintext in c for use
// during a session (PreRun decrypts them), so each is encrypted here into a local copy of its
// holder (never the live one) before being copied into the record; the copy's EncryptSecret /
// encryptStateTokensForContext call is a no-op if the live value is already ciphertext (e.g.
// when this runs inside save(), after save's own in-place encryption), so this is safe to call
// with either representation. A saved password and both nested API-key maps (GlobalAPIKeys,
// KafkaClusterConfig.APIKeys) are encrypted the moment they're stored and never decrypted back
// into the live config, so they cross here as an intact ciphertext/salt/nonce unit - no
// encrypt/decrypt.
func (c *Config) saveSecretStore() error {
	records := map[string]*secretRecord{}
	record := func(key string) *secretRecord {
		if r, ok := records[key]; ok {
			return r
		}
		r := &secretRecord{}
		records[key] = r
		return r
	}

	tokens := map[string]*tokenRecord{}
	for _, ctx := range c.Contexts {
		state := ctx.GetState()
		if state == nil || (state.AuthToken == "" && state.AuthRefreshToken == "") {
			continue
		}
		shadow := &Context{Name: ctx.Name, State: &ContextState{
			AuthToken:        state.AuthToken,
			AuthRefreshToken: state.AuthRefreshToken,
			Salt:             state.Salt,
			Nonce:            state.Nonce,
		}}
		if err := c.encryptStateTokensForContext(shadow, shadow.State.AuthToken, shadow.State.AuthRefreshToken); err != nil {
			return err
		}
		tokens[ctx.Name] = &tokenRecord{
			AuthToken:        shadow.State.AuthToken,
			AuthRefreshToken: shadow.State.AuthRefreshToken,
			Salt:             shadow.State.Salt,
			Nonce:            shadow.State.Nonce,
		}
	}

	for name, credential := range c.Credentials {
		if credential == nil || credential.APIKeyPair == nil || credential.APIKeyPair.Secret == "" {
			continue
		}
		shadow := &APIKeyPair{
			Key:    credential.APIKeyPair.Key,
			Secret: credential.APIKeyPair.Secret,
			Salt:   credential.APIKeyPair.Salt,
			Nonce:  credential.APIKeyPair.Nonce,
		}
		if err := shadow.EncryptSecret(); err != nil {
			return err
		}
		r := record(name)
		r.Secret = shadow.Secret
		r.SecretSalt = shadow.Salt
		r.SecretNonce = shadow.Nonce
	}

	// SavedCredentials is keyed by context name, not identity, so it is mapped to the
	// owning context's identityKey here.
	for ctxName, ctx := range c.Contexts {
		saved, ok := c.SavedCredentials[ctxName]
		if !ok || saved == nil || saved.EncryptedPassword == "" {
			continue
		}
		r := record(ctx.identityKey())
		r.Password = saved.EncryptedPassword
		r.PasswordSalt = saved.Salt
		r.PasswordNonce = saved.Nonce
	}

	for _, ctx := range c.Contexts {
		if len(ctx.GlobalAPIKeys) > 0 {
			r := record(ctx.identityKey())
			for keyId, pair := range ctx.GlobalAPIKeys {
				triple, err := encryptedAPIKeySecret(pair)
				if err != nil {
					return err
				}
				if triple == nil {
					continue
				}
				if r.GlobalAPIKeys == nil {
					r.GlobalAPIKeys = map[string]*apiKeySecret{}
				}
				r.GlobalAPIKeys[keyId] = triple
			}
		}

		for clusterId, cluster := range allKafkaClusterConfigs(ctx.KafkaClusterContext) {
			if cluster == nil || len(cluster.APIKeys) == 0 {
				continue
			}
			var keys map[string]*apiKeySecret
			for keyId, pair := range cluster.APIKeys {
				triple, err := encryptedAPIKeySecret(pair)
				if err != nil {
					return err
				}
				if triple == nil {
					continue
				}
				if keys == nil {
					keys = map[string]*apiKeySecret{}
				}
				keys[keyId] = triple
			}
			if len(keys) == 0 {
				continue
			}
			r := record(ctx.identityKey())
			if r.KafkaAPIKeys == nil {
				r.KafkaAPIKeys = map[string]map[string]*apiKeySecret{}
			}
			r.KafkaAPIKeys[clusterId] = keys
		}
	}

	if len(records) == 0 && len(tokens) == 0 {
		return nil
	}

	return newSecretStore().write(&secretFile{Secrets: records, Tokens: tokens})
}

// readSecretFileFromDisk reads and unmarshals the secret store at path, mirroring
// readConfigFromDisk. A missing file is not an error - a fresh install has no secrets.json
// yet - and returns an empty secretFile so Load still succeeds.
func readSecretFileFromDisk(path string) (*secretFile, error) {
	input, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &secretFile{}, nil
		}
		return nil, fmt.Errorf("unable to read secret store %s: %w", path, err)
	}

	file := &secretFile{}
	if err := json.Unmarshal(input, file); err != nil {
		return nil, fmt.Errorf("unable to unmarshal secret store %s: %w", path, err)
	}
	return file, nil
}

// loadSecretStore repopulates c's secret fields from the encrypted secret store - Secrets by
// each context's identityKey, Tokens by exact context name (no broadcast across contexts
// sharing an identity: see tokenRecord) - the reverse of saveSecretStore. It must run before
// wireContexts/Validate: ctx.GetState() is not wired to ContextStates until wireContexts
// runs, so this reads ContextStates/Credentials/SavedCredentials directly by name/identity;
// and Validate's nested-API-key pruning (validateGlobalAPIKeys, KafkaClusterContext.Validate)
// would delete a key whose secret still reads empty. Every value is copied verbatim - the
// ciphertext, salt, and nonce as one unit - never decrypted here; callers decrypt later via
// DecryptCredentials/DecryptContextStates/ResolveKafkaAPIKey.
func (c *Config) loadSecretStore() error {
	file, err := readSecretFileFromDisk(newSecretStore().path)
	if err != nil {
		return err
	}
	if len(file.Secrets) == 0 && len(file.Tokens) == 0 {
		return nil
	}

	for name, ctx := range c.Contexts {
		if tok, ok := file.Tokens[name]; ok && tok != nil && (tok.AuthToken != "" || tok.AuthRefreshToken != "") {
			if state := c.ContextStates[name]; state != nil {
				state.AuthToken = tok.AuthToken
				state.AuthRefreshToken = tok.AuthRefreshToken
				state.Salt = tok.Salt
				state.Nonce = tok.Nonce
			}
		}

		rec, ok := file.Secrets[ctx.identityKey()]
		if !ok || rec == nil {
			continue
		}

		if credential := c.Credentials[ctx.CredentialName]; credential != nil && credential.APIKeyPair != nil && rec.Secret != "" {
			credential.APIKeyPair.Secret = rec.Secret
			credential.APIKeyPair.Salt = rec.SecretSalt
			credential.APIKeyPair.Nonce = rec.SecretNonce
		}

		if saved := c.SavedCredentials[name]; saved != nil && rec.Password != "" {
			saved.EncryptedPassword = rec.Password
			saved.Salt = rec.PasswordSalt
			saved.Nonce = rec.PasswordNonce
		}

		for keyId, triple := range rec.GlobalAPIKeys {
			if pair := ctx.GlobalAPIKeys[keyId]; pair != nil && triple != nil {
				pair.Secret = triple.Secret
				pair.Salt = triple.Salt
				pair.Nonce = triple.Nonce
			}
		}

		clusters := allKafkaClusterConfigs(ctx.KafkaClusterContext)
		for clusterId, keys := range rec.KafkaAPIKeys {
			cluster := clusters[clusterId]
			if cluster == nil {
				continue
			}
			for keyId, triple := range keys {
				if pair := cluster.APIKeys[keyId]; pair != nil && triple != nil {
					pair.Secret = triple.Secret
					pair.Salt = triple.Salt
					pair.Nonce = triple.Nonce
				}
			}
		}
	}

	return nil
}
