package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// secretRecord holds one credential identity's secret material as stored on disk: each
// value is ciphertext paired with its own salt/nonce. A credential identity can be shared
// by several contexts (GlobalAPIKeys and nested Kafka API keys are identity-scoped), so each
// secret kind carries its own salt/nonce pair rather than sharing one - pairing a ciphertext
// with the wrong salt/nonce fails GCM authentication on load. Auth tokens and saved passwords
// are NOT identity-scoped: they are endpoint-scoped and keyed by context name instead, in
// tokenRecord/secretFile.Tokens and passwordRecord/secretFile.Passwords - see their doc comments.
type secretRecord struct {
	// Secret is the api-key secret, from APIKeyPair.
	Secret      string `json:"secret,omitempty"`
	SecretSalt  []byte `json:"secret_salt,omitempty"`
	SecretNonce []byte `json:"secret_nonce,omitempty"`

	// GlobalAPIKeys holds org-scoped API keys (Context.GlobalAPIKeys), keyed by API key id.
	GlobalAPIKeys map[string]*apiKeySecret `json:"global_api_keys,omitempty"`

	// KafkaAPIKeys holds per-cluster API keys (KafkaClusterConfig.APIKeys), keyed by cluster id
	// then by API key id.
	KafkaAPIKeys map[string]map[string]*apiKeySecret `json:"kafka_api_keys,omitempty"`

	// SchemaRegistryCredentials holds the (deprecated) SchemaRegistryCluster.SrCredentials secret,
	// keyed by Schema Registry cluster id. Unlike the Kafka/Global API keys, an SR credential is
	// never decrypted in place during a session, so it crosses the save as a stable ciphertext
	// triple and merges through the generic mergeMapDeep like Password.
	SchemaRegistryCredentials map[string]*apiKeySecret `json:"schema_registry_credentials,omitempty"`
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

// passwordRecord holds one context's saved login password as stored on disk, keyed by CONTEXT
// NAME (secretFile.Passwords), not credential identity. SavedCredentials is endpoint-specific: two
// contexts can share one credential identity (same username, different URL/CA-cert) while each
// holds its own saved password. Keying passwords by identity would collapse both into one record,
// so after reload both would get the last-writer password and one login would fail. The password
// is encrypted with the username as GCM associated data and is never decrypted in place during a
// session, so it crosses the save as a stable ciphertext triple.
type passwordRecord struct {
	Password string `json:"password,omitempty"`
	Salt     []byte `json:"salt,omitempty"`
	Nonce    []byte `json:"nonce,omitempty"`
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
// rename never orphans it. Tokens and Passwords are keyed by context name - see tokenRecord
// and passwordRecord.
type secretFile struct {
	Secrets   map[string]*secretRecord   `json:"secrets,omitempty"`
	Tokens    map[string]*tokenRecord    `json:"tokens,omitempty"`
	Passwords map[string]*passwordRecord `json:"passwords,omitempty"`
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

// stripSecretTriple returns a copy of rec with the CHURNING fields zeroed - the API-secret triple
// (Secret/SecretSalt/SecretNonce) and the nested API-key maps (GlobalAPIKeys, KafkaAPIKeys) - so the
// generic mergeMapDeep can safely structural-diff only what remains stable: SchemaRegistryCredentials
// (never decrypted in place, so its ciphertext is byte-stable across saves and platforms). The Secret
// triple and the nested maps are decrypted in place during a session (PreRun and ResolveKafkaAPIKey)
// and re-encrypted every save, which on Windows reproduces DIFFERENT ciphertext for an unchanged
// value; a byte-compare would misread that as a local change. Each is instead decided separately on
// PLAINTEXT (mergeSecretTriple, mergeAPIKeySecretMap) and spliced back in. SchemaRegistryCredentials
// stays present in the copy (a shared, read-only reference - mergeMapDeep only reads it via a JSON
// marshal). A nil rec copies to a non-nil empty record so an identity whose only content was a churning
// field still participates in the generic merge as "present, empty" rather than vanishing entirely.
func stripSecretTriple(rec *secretRecord) *secretRecord {
	if rec == nil {
		return &secretRecord{}
	}
	stripped := *rec
	stripped.Secret, stripped.SecretSalt, stripped.SecretNonce = "", nil, nil
	stripped.GlobalAPIKeys, stripped.KafkaAPIKeys = nil, nil
	return &stripped
}

// stripSecretTriples applies stripSecretTriple across a whole identity-keyed map.
func stripSecretTriples(records map[string]*secretRecord) map[string]*secretRecord {
	out := make(map[string]*secretRecord, len(records))
	for id, rec := range records {
		out[id] = stripSecretTriple(rec)
	}
	return out
}

// mergeSecretTriple three-way-merges one identity's API-secret (ciphertext, salt, nonce)
// unit. It cannot go through the generic mergeMapDeep like the rest of a secretRecord: a
// credential's API secret toggles to plaintext in the live config for use during a session
// (PreRun decrypts it) and is re-encrypted on every save, and on Windows that reproduces
// DIFFERENT ciphertext for an unchanged secret every time (DPAPI is non-deterministic and
// GenerateSaltAndNonce returns nil,nil there, so there is no salt/nonce to reuse the way
// AES-GCM does on Unix). A byte-level diff would then misread every untouched secret as
// locally changed on Windows - exactly the lost-write bug Config.decryptToMatch already
// solves at the config.json level, quoted in its own comment: re-encrypting to compare
// "would flag every secret as changed and reintroduce the very lost-write bug this guards
// against." The fix is the same here: decide on PLAINTEXT (decrypting base to compare,
// mirroring decryptToMatch), and move the whole (ciphertext, salt, nonce) triple as one
// unit so a decrypted comparison never ends up pairing one source's ciphertext with
// another's salt.
func (c *Config) mergeSecretTriple(identity string, base, ours, disk *secretRecord) (string, []byte, []byte, error) {
	credential := c.Credentials[identity]

	toPlain := func(rec *secretRecord) (string, error) {
		if rec == nil || rec.Secret == "" {
			return "", nil
		}
		if !isEncryptedSecret(rec.Secret) {
			return rec.Secret, nil
		}
		if credential == nil || credential.APIKeyPair == nil {
			// No key material to decrypt with: this process's live Credentials map has
			// nothing for this identity (only ours' emptiness matters for the branches
			// below, and ours is only ever non-empty when the identity IS in
			// c.Credentials, so this path is base-only). Read as "we have no opinion" -
			// not as an opaque non-empty value, which would wrongly look like "we hold a
			// secret" and block a legitimate concurrent add from disk.
			return "", nil
		}
		shadow := &APIKeyPair{Key: credential.APIKeyPair.Key, Secret: rec.Secret, Salt: rec.SecretSalt, Nonce: rec.SecretNonce}
		if err := shadow.DecryptSecret(); err != nil {
			return "", err
		}
		return shadow.Secret, nil
	}

	basePlain, err := toPlain(base)
	if err != nil {
		return "", nil, nil, err
	}
	oursPlain, err := toPlain(ours)
	if err != nil {
		return "", nil, nil, err
	}
	diskHas := disk != nil && disk.Secret != ""

	if oursPlain == "" {
		// We don't currently hold a secret for this identity. If we never did either, a
		// concurrent add on disk survives; if we used to (base had one), our clearing it
		// wins.
		if basePlain == "" && diskHas {
			return disk.Secret, disk.SecretSalt, disk.SecretNonce, nil
		}
		return "", nil, nil, nil
	}

	if oursPlain == basePlain {
		// Unchanged (by content, not by ciphertext bytes): defer to disk's current value,
		// exactly as if this process had never touched it. Preserves a concurrent rotation
		// this process never touched, and does not resurrect a concurrent clear.
		if diskHas {
			return disk.Secret, disk.SecretSalt, disk.SecretNonce, nil
		}
		return "", nil, nil, nil
	}

	// Changed (added or edited): our own value wins. It may still be plaintext here -
	// encryptSecrets/EncryptSecret has already run on it earlier in saveSecretStore, so in
	// practice this is ciphertext by the time it reaches here.
	return ours.Secret, ours.SecretSalt, ours.SecretNonce, nil
}

// decryptAPIKeySecret decrypts a nested API-key triple, using keyId as the encryption key: a nested
// key's map key IS its APIKeyPair.Key (StoreGlobalAPIKey/AddKafkaClusterConfig key each pair by its
// own Key). Empty or already-plaintext input is returned unchanged.
func decryptAPIKeySecret(keyId string, triple *apiKeySecret) (string, error) {
	if triple == nil || triple.Secret == "" {
		return "", nil
	}
	if !isEncryptedSecret(triple.Secret) {
		return triple.Secret, nil
	}
	shadow := &APIKeyPair{Key: keyId, Secret: triple.Secret, Salt: triple.Salt, Nonce: triple.Nonce}
	if err := shadow.DecryptSecret(); err != nil {
		return "", err
	}
	return shadow.Secret, nil
}

// mergeAPIKeySecretMap three-way-merges one nested API-key map (a context's GlobalAPIKeys, or a
// single cluster's slice of KafkaAPIKeys) per key on decrypted PLAINTEXT, the nested-key analog of
// mergeSecretTriple with the same Windows non-determinism hazard and fix: ResolveKafkaAPIKey decrypts
// these pairs in place for the session and they re-encrypt to different ciphertext every save, so a
// byte-compare would flag an untouched key as changed and clobber a concurrent rotation/deletion. The
// intact (ciphertext, salt, nonce) triple is moved as one unit so a plaintext decision never pairs one
// source's ciphertext with another's salt. Returns nil for an empty result so an identity contributes
// no empty map.
func (c *Config) mergeAPIKeySecretMap(base, ours, disk map[string]*apiKeySecret) (map[string]*apiKeySecret, error) {
	keyIds := map[string]bool{}
	for id := range base {
		keyIds[id] = true
	}
	for id := range ours {
		keyIds[id] = true
	}
	for id := range disk {
		keyIds[id] = true
	}

	merged := map[string]*apiKeySecret{}
	for keyId := range keyIds {
		basePlain, err := decryptAPIKeySecret(keyId, base[keyId])
		if err != nil {
			return nil, err
		}
		oursPlain, err := decryptAPIKeySecret(keyId, ours[keyId])
		if err != nil {
			return nil, err
		}
		diskHas := disk[keyId] != nil && disk[keyId].Secret != ""

		switch {
		case oursPlain == "":
			// We don't hold this key. A concurrent add on disk survives if we never did either;
			// our clearing it (base had one) wins otherwise.
			if basePlain == "" && diskHas {
				merged[keyId] = disk[keyId]
			}
		case oursPlain == basePlain:
			// Unchanged by content: defer to disk, preserving a concurrent rotation this process
			// never touched and not resurrecting a concurrent delete.
			if diskHas {
				merged[keyId] = disk[keyId]
			}
		default:
			// Added or edited: our own value wins.
			merged[keyId] = ours[keyId]
		}
	}

	if len(merged) == 0 {
		return nil, nil
	}
	return merged, nil
}

// mergeKafkaAPIKeys three-way-merges the cluster-keyed KafkaAPIKeys map by delegating each cluster's
// inner key map to mergeAPIKeySecretMap. Returns nil for an empty result.
func (c *Config) mergeKafkaAPIKeys(base, ours, disk map[string]map[string]*apiKeySecret) (map[string]map[string]*apiKeySecret, error) {
	clusterIds := map[string]bool{}
	for id := range base {
		clusterIds[id] = true
	}
	for id := range ours {
		clusterIds[id] = true
	}
	for id := range disk {
		clusterIds[id] = true
	}

	merged := map[string]map[string]*apiKeySecret{}
	for clusterId := range clusterIds {
		keys, err := c.mergeAPIKeySecretMap(base[clusterId], ours[clusterId], disk[clusterId])
		if err != nil {
			return nil, err
		}
		if len(keys) > 0 {
			merged[clusterId] = keys
		}
	}

	if len(merged) == 0 {
		return nil, nil
	}
	return merged, nil
}

// mergeToken three-way-merges one context's auth-token (ciphertext, salt, nonce) unit - the
// token analog of mergeSecretTriple, with the same Windows non-determinism hazard (a
// current context's tokens toggle to plaintext for the session and are re-encrypted every
// save) and the same fix. Unlike a secretRecord, a tokenRecord is ENTIRELY the churning
// unit - Tokens is never run through the generic mergeMapDeep at all; this function alone
// decides each context's whole entry, including add/delete.
func (c *Config) mergeToken(ctxName string, base, ours, disk *tokenRecord) (string, string, []byte, []byte, error) {
	toPlain := func(rec *tokenRecord) (string, string, error) {
		if rec == nil {
			return "", "", nil
		}
		shadow := &ContextState{AuthToken: rec.AuthToken, AuthRefreshToken: rec.AuthRefreshToken, Salt: rec.Salt, Nonce: rec.Nonce}
		if isEncryptedSecret(shadow.AuthToken) {
			if err := shadow.DecryptAuthToken(ctxName); err != nil {
				return "", "", err
			}
		}
		if isEncryptedSecret(shadow.AuthRefreshToken) {
			if err := shadow.DecryptAuthRefreshToken(ctxName); err != nil {
				return "", "", err
			}
		}
		return shadow.AuthToken, shadow.AuthRefreshToken, nil
	}

	baseToken, baseRefresh, err := toPlain(base)
	if err != nil {
		return "", "", nil, nil, err
	}
	oursToken, oursRefresh, err := toPlain(ours)
	if err != nil {
		return "", "", nil, nil, err
	}
	diskHas := disk != nil && (disk.AuthToken != "" || disk.AuthRefreshToken != "")

	if oursToken == "" && oursRefresh == "" {
		if baseToken == "" && baseRefresh == "" && diskHas {
			return disk.AuthToken, disk.AuthRefreshToken, disk.Salt, disk.Nonce, nil
		}
		return "", "", nil, nil, nil
	}

	if oursToken == baseToken && oursRefresh == baseRefresh {
		if diskHas {
			return disk.AuthToken, disk.AuthRefreshToken, disk.Salt, disk.Nonce, nil
		}
		return "", "", nil, nil, nil
	}

	return ours.AuthToken, ours.AuthRefreshToken, ours.Salt, ours.Nonce, nil
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
// via saveSecretStore(c).
//
// It sets presence only (any non-empty value; merged's own Secret field is json:"-" and never
// reaches the write either way), from two sources: (1) this process's own live c, for a key c added
// this session that is not yet on disk; and (2) the on-disk secret store, for a key ONLY a concurrent
// session added - such a key rides in via the config.json merge (its public id in disk -> merged) but
// with an empty secret, and c has never heard of it, so without this pass Validate() would prune its
// id before the write. saveSecretStore's own three-way merge then reconciles the secret material
// itself; this function only keeps Validate() from pruning the (json:"-", so structurally invisible)
// public id in the meantime.
func rehydrateNestedAPIKeySecretPresence(c, merged *Config) error {
	// Pass 1: keys this process's own live c knows about (including ones not yet persisted).
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

	// Pass 2: disk-only keys a concurrent session added, still empty in merged after pass 1.
	disk, err := readSecretFileFromDisk(newSecretStore().path)
	if err != nil {
		return err
	}
	for _, mergedCtx := range merged.Contexts {
		if mergedCtx == nil {
			continue
		}
		rec := disk.Secrets[mergedCtx.identityKey()]
		if rec == nil {
			continue
		}

		for keyId, mergedPair := range mergedCtx.GlobalAPIKeys {
			if mergedPair == nil || mergedPair.Secret != "" {
				continue
			}
			if triple := rec.GlobalAPIKeys[keyId]; triple != nil && triple.Secret != "" {
				mergedPair.Secret = triple.Secret
			}
		}

		for clusterId, cluster := range allKafkaClusterConfigs(mergedCtx.KafkaClusterContext) {
			keys := rec.KafkaAPIKeys[clusterId]
			if keys == nil {
				continue
			}
			for keyId, mergedPair := range cluster.APIKeys {
				if mergedPair == nil || mergedPair.Secret != "" {
					continue
				}
				if triple := keys[keyId]; triple != nil && triple.Secret != "" {
					mergedPair.Secret = triple.Secret
				}
			}
		}
	}

	return nil
}

// saveSecretStore extracts c's secret material into the secret store, keyed by identity
// (Secrets) or by context name (Tokens), and persists it under the caller's already-held
// save lock. It reads the live c directly, never a merged/deep-copied Config: every secret
// field is json:"-", so json.Marshal (the copy mechanism threeWayMerge's inputs go through)
// silently drops it, and a value read back from that copy is always empty - encrypting that
// empty value would quietly replace the real secret with ciphertext of "".
//
// A credential's API secret and a context's auth tokens toggle to plaintext in c for use
// during a session (PreRun decrypts them), so each is encrypted here into a local copy of its
// holder (never the live one) before being copied into the record; the copy's EncryptSecret /
// encryptStateTokensForContext call is a no-op if the live value is already ciphertext (e.g.
// when this runs inside save(), after save's own in-place encryption), so this is safe to call
// with either representation. On Unix this reproduces byte-identical ciphertext for an
// untouched value (salt/nonce are reused whenever already set), but on Windows it does NOT:
// DPAPI is non-deterministic and GenerateSaltAndNonce returns nil,nil there, so a plaintext
// value re-encrypts to different ciphertext every save. mergeSecretTriple/mergeToken below
// diff on PLAINTEXT specifically to tolerate that; see their own comments. A saved password
// and both nested API-key maps (GlobalAPIKeys, KafkaClusterConfig.APIKeys) are encrypted the
// moment they're stored and never decrypted back into the live config, so they cross here as
// an intact ciphertext/salt/nonce unit that IS stable across platforms - encryptedAPIKeySecret
// (and Password's direct copy below) never re-derives them, so the generic mergeMapDeep is
// safe for those two.
//
// diskContextNames is the set of context names present in config.json as read from disk this
// same save cycle (nil when the caller is writing a whole config with nothing to merge
// against, e.g. a fresh install or a from-scratch write): see the three-way-merge branch below.
func (c *Config) saveSecretStore(diskContextNames map[string]bool) error {
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

	// Saved passwords are endpoint-specific (SavedCredentials is keyed by context name), so they
	// are stored by context name in their own map - never folded into the identity-keyed record,
	// where two contexts sharing one credential identity would collide. See passwordRecord.
	passwords := map[string]*passwordRecord{}
	for ctxName := range c.Contexts {
		saved, ok := c.SavedCredentials[ctxName]
		if !ok || saved == nil || saved.EncryptedPassword == "" {
			continue
		}
		passwords[ctxName] = &passwordRecord{
			Password: saved.EncryptedPassword,
			Salt:     saved.Salt,
			Nonce:    saved.Nonce,
		}
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

		for srClusterId, srCluster := range ctx.SchemaRegistryClusters {
			if srCluster == nil {
				continue
			}
			triple, err := encryptedAPIKeySecret(srCluster.SrCredentials)
			if err != nil {
				return err
			}
			if triple == nil {
				continue
			}
			r := record(ctx.identityKey())
			if r.SchemaRegistryCredentials == nil {
				r.SchemaRegistryCredentials = map[string]*apiKeySecret{}
			}
			r.SchemaRegistryCredentials[srClusterId] = triple
		}
	}

	ours := &secretFile{Secrets: records, Tokens: tokens, Passwords: passwords}

	// No baseline (never loaded) or no disk read (the caller is writing a whole config with
	// nothing to merge against) means there is no common ancestor: declare our state whole,
	// matching saveLocked's own nil-baseline/missing-file short circuits for config.json.
	if diskContextNames == nil || c.secretBaseline == nil {
		if len(records) == 0 && len(tokens) == 0 && len(passwords) == 0 {
			return nil
		}
		if err := newSecretStore().write(ours); err != nil {
			return err
		}
		c.secretBaseline = ours
		return nil
	}

	store := newSecretStore()
	disk, err := readSecretFileFromDisk(store.path)
	if err != nil {
		return err
	}

	// Secrets: the generic structural merge (identity add/delete, SchemaRegistryCredentials) runs on
	// copies with the churning fields - the API-secret triple and the nested API-key maps - stripped
	// out, so their platform-dependent (re-)encryption never contaminates it (see stripSecretTriple).
	// Each churning field is decided separately, on plaintext (mergeSecretTriple for the credential
	// secret, mergeAPIKeySecretMap/mergeKafkaAPIKeys for the nested keys), and spliced back in below.
	merged := &secretFile{}
	if merged.Secrets, err = mergeMapDeep(
		stripSecretTriples(c.secretBaseline.Secrets), stripSecretTriples(records), stripSecretTriples(disk.Secrets),
	); err != nil {
		return fmt.Errorf("unable to merge secret store: %w", err)
	}

	identities := map[string]bool{}
	for id := range c.secretBaseline.Secrets {
		identities[id] = true
	}
	for id := range records {
		identities[id] = true
	}
	for id := range disk.Secrets {
		identities[id] = true
	}
	globalOf := func(rec *secretRecord) map[string]*apiKeySecret {
		if rec == nil {
			return nil
		}
		return rec.GlobalAPIKeys
	}
	kafkaOf := func(rec *secretRecord) map[string]map[string]*apiKeySecret {
		if rec == nil {
			return nil
		}
		return rec.KafkaAPIKeys
	}
	for id := range identities {
		base, ours, dsk := c.secretBaseline.Secrets[id], records[id], disk.Secrets[id]

		secret, salt, nonce, err := c.mergeSecretTriple(id, base, ours, dsk)
		if err != nil {
			return err
		}
		global, err := c.mergeAPIKeySecretMap(globalOf(base), globalOf(ours), globalOf(dsk))
		if err != nil {
			return err
		}
		kafka, err := c.mergeKafkaAPIKeys(kafkaOf(base), kafkaOf(ours), kafkaOf(dsk))
		if err != nil {
			return err
		}

		rec, ok := merged.Secrets[id]
		if !ok {
			if secret == "" && len(global) == 0 && len(kafka) == 0 {
				continue
			}
			rec = &secretRecord{}
			merged.Secrets[id] = rec
		}
		rec.Secret, rec.SecretSalt, rec.SecretNonce = secret, salt, nonce
		rec.GlobalAPIKeys, rec.KafkaAPIKeys = global, kafka
	}
	// Drop an identity left with no content at all: its only material was a churning field,
	// and that's now cleared.
	for id, rec := range merged.Secrets {
		if rec.Secret == "" && len(rec.GlobalAPIKeys) == 0 && len(rec.KafkaAPIKeys) == 0 && len(rec.SchemaRegistryCredentials) == 0 {
			delete(merged.Secrets, id)
		}
	}

	// Tokens: a tokenRecord is entirely the churning unit (unlike a stable password or the nested
	// key/SR triples), so mergeToken alone decides each context's whole entry - no generic
	// mergeMapDeep pass first.
	merged.Tokens = map[string]*tokenRecord{}
	ctxNames := map[string]bool{}
	for name := range c.secretBaseline.Tokens {
		ctxNames[name] = true
	}
	for name := range tokens {
		ctxNames[name] = true
	}
	for name := range disk.Tokens {
		ctxNames[name] = true
	}
	for name := range ctxNames {
		authToken, authRefreshToken, salt, nonce, err := c.mergeToken(name, c.secretBaseline.Tokens[name], tokens[name], disk.Tokens[name])
		if err != nil {
			return err
		}
		if authToken == "" && authRefreshToken == "" {
			continue
		}
		merged.Tokens[name] = &tokenRecord{AuthToken: authToken, AuthRefreshToken: authRefreshToken, Salt: salt, Nonce: nonce}
	}

	// A token's presence is coupled to its owning context surviving, exactly like
	// ContextStates is coupled to Contexts in threeWayMerge. mergeToken alone cannot tell a
	// deliberate token clear (context stays; only the token empties, e.g. logout - must not be
	// resurrected) from a token that only vanished from disk because a CONCURRENT session
	// deleted its owning context, which this process's own edit then revives (that token must
	// not be lost). The distinguishing signal is whether the context was already on disk: if
	// it was, an untouched-but-now-missing token is a deliberate clear elsewhere, so mergeToken
	// correctly dropped it above; if it wasn't, this save is the one reviving the context, so
	// its token comes along too.
	for name, tok := range tokens {
		if _, ok := merged.Tokens[name]; ok {
			continue
		}
		if diskContextNames[name] {
			continue
		}
		merged.Tokens[name] = tok
	}

	// Passwords: a saved password is a stable ciphertext triple (never decrypted in place during a
	// session, so no Windows re-encryption churn), so unlike tokens it merges through the generic
	// structural mergeMapDeep - the same three-way merge SavedCredentials itself uses in config.json.
	if merged.Passwords, err = mergeMapDeep(c.secretBaseline.Passwords, passwords, disk.Passwords); err != nil {
		return fmt.Errorf("unable to merge secret store passwords: %w", err)
	}

	if err := store.write(merged); err != nil {
		return err
	}
	c.secretBaseline = ours
	return nil
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
// each context's identityKey, Tokens and Passwords by exact context name (no broadcast across
// contexts sharing an identity: see tokenRecord/passwordRecord) - the reverse of saveSecretStore.
// It must run before
// wireContexts/Validate: ctx.GetState() is not wired to ContextStates until wireContexts
// runs, so this reads ContextStates/Credentials/SavedCredentials directly by name/identity;
// and Validate's nested-API-key pruning (validateGlobalAPIKeys, KafkaClusterContext.Validate)
// would delete a key whose secret still reads empty. Every value is copied verbatim - the
// ciphertext, salt, and nonce as one unit - never decrypted here; callers decrypt later via
// DecryptCredentials/DecryptContextStates/ResolveKafkaAPIKey.
//
// The file read here is also snapshotted as c.secretBaseline, this process's common ancestor
// for saveSecretStore's own three-way merge - mirroring snapshotBaseline for config.json,
// except the secret store needs a separate baseline because deepCopyPersisted's JSON round
// trip drops every json:"-" secret field and so cannot hold one.
func (c *Config) loadSecretStore() error {
	file, err := readSecretFileFromDisk(newSecretStore().path)
	if err != nil {
		return err
	}
	c.secretBaseline = file
	if len(file.Secrets) == 0 && len(file.Tokens) == 0 && len(file.Passwords) == 0 {
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

		// Passwords are keyed by context name, not identity (see passwordRecord), so they are
		// repopulated by exact context name - independent of the identity-keyed secretRecord, which
		// may be absent for a context whose only stored secret is a password.
		if pw, ok := file.Passwords[name]; ok && pw != nil && pw.Password != "" {
			if saved := c.SavedCredentials[name]; saved != nil {
				saved.EncryptedPassword = pw.Password
				saved.Salt = pw.Salt
				saved.Nonce = pw.Nonce
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

		for srClusterId, triple := range rec.SchemaRegistryCredentials {
			srCluster := ctx.SchemaRegistryClusters[srClusterId]
			if srCluster == nil || srCluster.SrCredentials == nil || triple == nil {
				continue
			}
			srCluster.SrCredentials.Secret = triple.Secret
			srCluster.SrCredentials.Salt = triple.Salt
			srCluster.SrCredentials.Nonce = triple.Nonce
		}
	}

	return nil
}
