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

// stripSecretTriple returns a copy of rec with Secret/SecretSalt/SecretNonce zeroed, so the
// generic mergeMapDeep can safely structural-diff everything else in a secretRecord
// (Password, GlobalAPIKeys, KafkaAPIKeys - all stable ciphertext, see mergeSecretTriple's
// comment) without the churning API-secret triple contaminating that decision. Password and
// the nested API-key maps are still present in the copy (shared, read-only references -
// mergeMapDeep only ever reads them via a JSON marshal). A nil rec copies to a non-nil empty
// record so an identity whose only content was the secret still participates in the
// generic merge as "present, empty" rather than vanishing from the key set entirely.
func stripSecretTriple(rec *secretRecord) *secretRecord {
	if rec == nil {
		return &secretRecord{}
	}
	stripped := *rec
	stripped.Secret, stripped.SecretSalt, stripped.SecretNonce = "", nil, nil
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
func (c *Config) mergeSecretTriple(identity string, base, ours, disk *secretRecord) (secret string, salt, nonce []byte, err error) {
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

// mergeToken three-way-merges one context's auth-token (ciphertext, salt, nonce) unit - the
// token analog of mergeSecretTriple, with the same Windows non-determinism hazard (a
// current context's tokens toggle to plaintext for the session and are re-encrypted every
// save) and the same fix. Unlike a secretRecord, a tokenRecord is ENTIRELY the churning
// unit - Tokens is never run through the generic mergeMapDeep at all; this function alone
// decides each context's whole entry, including add/delete.
func (c *Config) mergeToken(ctxName string, base, ours, disk *tokenRecord) (authToken, authRefreshToken string, salt, nonce []byte, err error) {
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
// via saveSecretStore(c). This sets presence only (any non-empty value; merged's own Secret
// field is json:"-" and never reaches the write below either way) for a key this process's own
// live c already knows about. A key only a concurrent session added exists only on disk and
// is not rehydrated here - saveSecretStore's own three-way merge (mergeMapDeep over
// GlobalAPIKeys/KafkaAPIKeys) reconciles that at the secret-store level; this function only
// keeps Validate() from pruning the (json:"-", so structurally invisible to it) key's public id
// in the meantime.
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

	ours := &secretFile{Secrets: records, Tokens: tokens}

	// No baseline (never loaded) or no disk read (the caller is writing a whole config with
	// nothing to merge against) means there is no common ancestor: declare our state whole,
	// matching saveLocked's own nil-baseline/missing-file short circuits for config.json.
	if diskContextNames == nil || c.secretBaseline == nil {
		if len(records) == 0 && len(tokens) == 0 {
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

	// Secrets: the generic structural merge (identity add/delete, Password, the nested
	// API-key maps) runs on copies with the churning API-secret triple stripped out, so that
	// triple's platform-dependent (re-)encryption never contaminates it - see
	// stripSecretTriple. The triple itself is decided separately, on plaintext, by
	// mergeSecretTriple and spliced back in below.
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
	for id := range identities {
		secret, salt, nonce, err := c.mergeSecretTriple(id, c.secretBaseline.Secrets[id], records[id], disk.Secrets[id])
		if err != nil {
			return err
		}
		rec, ok := merged.Secrets[id]
		if !ok {
			if secret == "" {
				continue
			}
			rec = &secretRecord{}
			merged.Secrets[id] = rec
		}
		rec.Secret, rec.SecretSalt, rec.SecretNonce = secret, salt, nonce
	}
	// Drop an identity left with no content at all: its only material was the secret
	// triple, and that's now cleared.
	for id, rec := range merged.Secrets {
		if rec.Secret == "" && rec.Password == "" && len(rec.GlobalAPIKeys) == 0 && len(rec.KafkaAPIKeys) == 0 {
			delete(merged.Secrets, id)
		}
	}

	// Tokens: a tokenRecord is entirely the churning unit (no stable sibling fields the way
	// Password/nested keys are for a secretRecord), so mergeToken alone decides each
	// context's whole entry - no generic mergeMapDeep pass first.
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
// each context's identityKey, Tokens by exact context name (no broadcast across contexts
// sharing an identity: see tokenRecord) - the reverse of saveSecretStore. It must run before
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
