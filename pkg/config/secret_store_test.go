package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/secret"
)

// TestSave_SecretsLeaveConfigFile pins the extraction contract: a saved API-key secret,
// context auth tokens, and a saved login password must not appear in config.json
// (plaintext or ciphertext), and must each decrypt back to their real value from the
// secret store - not just be absent from either file as ciphertext-shaped noise. A
// ciphertext of "" (this test's regression: see task-2.2-report.md) passes every
// absence check below but fails every decrypt-back check.
func TestSave_SecretsLeaveConfigFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	ctx := c.Contexts["orig"]

	ctx.State.AuthToken = "header.payload.signature"
	ctx.State.AuthRefreshToken = "v1.some-refresh-token"

	salt, nonce, err := secret.GenerateSaltAndNonce()
	require.NoError(t, err)
	encryptedPassword, err := secret.Encrypt("orig-user", "the-password", salt, nonce)
	require.NoError(t, err)
	c.SavedCredentials["orig"] = &LoginCredential{
		Username:          "orig-user",
		EncryptedPassword: encryptedPassword,
		Salt:              salt,
		Nonce:             nonce,
	}

	require.NoError(t, c.Save())

	cfgRaw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	require.NotContains(t, string(cfgRaw), "the-api-secret")
	require.NotContains(t, string(cfgRaw), `"secret"`)
	require.NotContains(t, string(cfgRaw), "header.payload.signature")
	require.NotContains(t, string(cfgRaw), "the-password")

	secRaw, err := os.ReadFile(SecretsFilename())
	require.NoError(t, err)
	require.Contains(t, string(secRaw), "secrets")
	require.Contains(t, string(secRaw), "tokens")
	require.NotContains(t, string(secRaw), "the-api-secret") // encrypted, not plaintext
	require.NotContains(t, string(secRaw), "header.payload.signature")
	require.NotContains(t, string(secRaw), "the-password")

	var file secretFile
	require.NoError(t, json.Unmarshal(secRaw, &file))
	rec := file.Secrets["api-key-AK"]
	require.NotNil(t, rec)

	plainSecret, err := secret.Decrypt("AK", rec.Secret, rec.SecretSalt, rec.SecretNonce)
	require.NoError(t, err)
	require.Equal(t, "the-api-secret", plainSecret)

	tok := file.Tokens["orig"]
	require.NotNil(t, tok)

	plainAuthToken, err := secret.Decrypt("orig", tok.AuthToken, tok.Salt, tok.Nonce)
	require.NoError(t, err)
	require.Equal(t, "header.payload.signature", plainAuthToken)

	plainRefreshToken, err := secret.Decrypt("orig", tok.AuthRefreshToken, tok.Salt, tok.Nonce)
	require.NoError(t, err)
	require.Equal(t, "v1.some-refresh-token", plainRefreshToken)

	pw := file.Passwords["orig"]
	require.NotNil(t, pw)
	plainPassword, err := secret.Decrypt("orig-user", pw.Password, pw.Salt, pw.Nonce)
	require.NoError(t, err)
	require.Equal(t, "the-password", plainPassword)
}

// TestSave_NestedApiKeySecretsLeaveConfigFile pins the extraction contract for nested
// API-key secrets: a global or cluster-scoped key's id stays in config.json (public
// metadata), but its secret leaves for the secret store and must decrypt back to its
// real value there, not just be absent from either file as ciphertext-shaped noise.
func TestSave_NestedApiKeySecretsLeaveConfigFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	ctx := c.Contexts["orig"]

	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))

	cluster := &KafkaClusterConfig{
		ID:        "lkc-nested",
		Name:      "nested-cluster",
		Bootstrap: "https://nested.example.com",
		APIKeys:   map[string]*APIKeyPair{"CLUSTER-KEY": {Key: "CLUSTER-KEY", Secret: "cluster-secret"}},
	}
	ctx.KafkaClusterContext.AddKafkaClusterConfig(cluster)
	require.NoError(t, cluster.EncryptAPIKeys())

	require.NoError(t, c.Save())

	cfgRaw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	require.Contains(t, string(cfgRaw), `"GLOBAL-KEY"`)
	require.Contains(t, string(cfgRaw), `"CLUSTER-KEY"`)
	require.NotContains(t, string(cfgRaw), "global-secret")
	require.NotContains(t, string(cfgRaw), "cluster-secret")

	secRaw, err := os.ReadFile(SecretsFilename())
	require.NoError(t, err)
	require.Contains(t, string(secRaw), "global_api_keys")
	require.Contains(t, string(secRaw), "kafka_api_keys")
	require.NotContains(t, string(secRaw), "global-secret")
	require.NotContains(t, string(secRaw), "cluster-secret")

	var file secretFile
	require.NoError(t, json.Unmarshal(secRaw, &file))
	rec := file.Secrets["api-key-AK"]
	require.NotNil(t, rec)

	plainSecret, err := secret.Decrypt("AK", rec.Secret, rec.SecretSalt, rec.SecretNonce)
	require.NoError(t, err)
	require.Equal(t, "the-api-secret", plainSecret)

	globalTriple := rec.GlobalAPIKeys["GLOBAL-KEY"]
	require.NotNil(t, globalTriple)
	plainGlobal, err := secret.Decrypt("GLOBAL-KEY", globalTriple.Secret, globalTriple.Salt, globalTriple.Nonce)
	require.NoError(t, err)
	require.Equal(t, "global-secret", plainGlobal)

	clusterTriple := rec.KafkaAPIKeys["lkc-nested"]["CLUSTER-KEY"]
	require.NotNil(t, clusterTriple)
	plainCluster, err := secret.Decrypt("CLUSTER-KEY", clusterTriple.Secret, clusterTriple.Salt, clusterTriple.Nonce)
	require.NoError(t, err)
	require.Equal(t, "cluster-secret", plainCluster)
}

// TestLoad_RepopulatesSecretsFromStore pins the load-side counterpart of
// TestSave_SecretsLeaveConfigFile: a credential's API secret, extracted to the secret
// store on save, must be repopulated into a freshly loaded Config so DecryptCredentials
// still recovers the real plaintext.
func TestLoad_RepopulatesSecretsFromStore(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())
	require.NoError(t, reloaded.DecryptCredentials())

	require.Equal(t, "the-api-secret", reloaded.Credentials["api-key-AK"].APIKeyPair.Secret)
}

// TestSecrets_SurviveContextRename pins that the secret store's identity key
// (Context.identityKey, CredentialName) stays stable across a rename, so a renamed
// context's secrets are not orphaned in the store.
func TestSecrets_SurviveContextRename(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	require.NoError(t, c.Save())
	require.NoError(t, c.renameContextForTest("orig", "renamed"))
	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())
	require.NoError(t, reloaded.DecryptCredentials())
	require.Equal(t, "the-api-secret", reloaded.Credentials["api-key-AK"].APIKeyPair.Secret)
}

// TestLoad_DecryptsOverEncryptedCloudNonV1RefreshToken covers the deferred edge from the
// task-2.2 review: saveSecretStore's shadow context (used to encrypt state tokens) has no
// PlatformName, so ctx.IsCloud reads false inside encryptStateTokensForContext even for a
// real cloud context, and a cloud refresh token that isn't `v1.`-prefixed gets encrypted
// there when it otherwise might not be. The stored ciphertext is still self-describing
// (cipher prefix + TokenSalt/Nonce), so it must decrypt back to the original token on load
// regardless of why it was encrypted.
func TestLoad_DecryptsOverEncryptedCloudNonV1RefreshToken(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	ctx := c.Contexts["orig"]

	cloudPlatform := &Platform{Name: "confluent.cloud", Server: "https://confluent.cloud"}
	c.Platforms["confluent.cloud"] = cloudPlatform
	ctx.PlatformName = "confluent.cloud"
	ctx.Platform = cloudPlatform
	require.True(t, ctx.IsCloud(false), "precondition: the real context must read as cloud")

	ctx.State.AuthRefreshToken = "opaque-cloud-refresh-token" // cloud token, not v1.-prefixed
	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())

	reloadedCtx := reloaded.Contexts["orig"]
	require.NoError(t, reloadedCtx.GetState().DecryptAuthRefreshToken(reloadedCtx.Name))
	require.Equal(t, "opaque-cloud-refresh-token", reloadedCtx.GetState().AuthRefreshToken)
}

// TestLoad_RepopulatesNestedAPIKeySecretsFromStore is the load-side counterpart of
// TestSave_NestedApiKeySecretsLeaveConfigFile: a global API key's and a cluster-scoped API
// key's secrets, extracted to the secret store on save, must be repopulated into a freshly
// loaded Config so each decrypts back to its real value.
func TestLoad_RepopulatesNestedAPIKeySecretsFromStore(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	ctx := c.Contexts["orig"]

	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GLOBAL-KEY", Secret: "global-secret"}))

	cluster := &KafkaClusterConfig{
		ID:        "lkc-nested",
		Name:      "nested-cluster",
		Bootstrap: "https://nested.example.com",
		APIKeys:   map[string]*APIKeyPair{"CLUSTER-KEY": {Key: "CLUSTER-KEY", Secret: "cluster-secret"}},
	}
	ctx.KafkaClusterContext.AddKafkaClusterConfig(cluster)
	require.NoError(t, cluster.EncryptAPIKeys())

	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())

	reloadedCtx := reloaded.Contexts["orig"]
	require.NoError(t, reloadedCtx.DecryptGlobalAPIKeys())
	require.Equal(t, "global-secret", reloadedCtx.GlobalAPIKeys["GLOBAL-KEY"].Secret)

	reloadedCluster := reloadedCtx.KafkaClusterContext.KafkaClusterConfigs["lkc-nested"]
	require.NoError(t, reloadedCluster.DecryptAPIKeys())
	require.Equal(t, "cluster-secret", reloadedCluster.APIKeys["CLUSTER-KEY"].Secret)
}

// TestSave_EmptyCredentialSecretNotEncrypted is the Unix-simulated counterpart of the Windows
// DPAPI failure: encryptCredentialsAPISecret runs over every credential on save, and a
// credential with an empty API secret (e.g. a username/password login) must not be handed to
// secret.Encrypt - DPAPI rejects empty input on Windows. The empty secret must stay empty and
// produce no ciphertext-of-"" in either file.
func TestSave_EmptyCredentialSecretNotEncrypted(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	c.Credentials["empty-cred"] = &Credential{
		Name:           "empty-cred",
		CredentialType: Username,
		APIKeyPair:     &APIKeyPair{},
	}

	require.NoError(t, c.Save())

	require.Empty(t, c.Credentials["empty-cred"].APIKeyPair.Secret,
		"an empty credential secret must stay empty, never encrypted into ciphertext-of-empty")

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())
	require.NoError(t, reloaded.DecryptCredentials())
	require.Empty(t, reloaded.Credentials["empty-cred"].APIKeyPair.Secret)
}

// TestEncryptStateTokensForContext_EmptyGovRefreshTokenNotEncrypted pins the token analog of the
// empty-encrypt guard: a Confluent Gov context with no refresh token must not encrypt the empty
// string (the gov branch does not otherwise require a non-empty token), which DPAPI would reject.
func TestEncryptStateTokensForContext_EmptyGovRefreshTokenNotEncrypted(t *testing.T) {
	c := New()
	ctx := &Context{Name: "gov", PlatformName: "confluentgov.com", State: &ContextState{}}

	require.NoError(t, c.encryptStateTokensForContext(ctx, "", ""))

	require.Empty(t, ctx.State.AuthToken)
	require.Empty(t, ctx.State.AuthRefreshToken, "an empty gov refresh token must not be encrypted")
}

// TestSave_SchemaRegistryCredentialLeavesConfigFile pins that a (deprecated) Schema Registry
// cluster's API secret leaves config.json for the secret store on save and decrypts back to its
// real value there. SrCredentials is the fourth *APIKeyPair holder; with APIKeyPair.Secret retagged
// json:"-", an unhandled SrCredentials secret would be dropped from config.json and stored nowhere.
func TestSave_SchemaRegistryCredentialLeavesConfigFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	ctx := c.Contexts["orig"]
	ctx.SchemaRegistryClusters = map[string]*SchemaRegistryCluster{
		"lsrc-1": {
			Id:                     "lsrc-1",
			SchemaRegistryEndpoint: "https://sr.example.com",
			SrCredentials:          &APIKeyPair{Key: "SR-KEY", Secret: "sr-secret"},
		},
	}

	require.NoError(t, c.Save())

	cfgRaw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	require.Contains(t, string(cfgRaw), `"SR-KEY"`)
	require.NotContains(t, string(cfgRaw), "sr-secret")

	secRaw, err := os.ReadFile(SecretsFilename())
	require.NoError(t, err)
	require.NotContains(t, string(secRaw), "sr-secret")

	var file secretFile
	require.NoError(t, json.Unmarshal(secRaw, &file))
	triple := file.Secrets["api-key-AK"].SchemaRegistryCredentials["lsrc-1"]
	require.NotNil(t, triple)
	plain, err := secret.Decrypt("SR-KEY", triple.Secret, triple.Salt, triple.Nonce)
	require.NoError(t, err)
	require.Equal(t, "sr-secret", plain)
}

// TestLoad_RepopulatesSchemaRegistryCredentialFromStore is the load-side counterpart: a Schema
// Registry credential extracted to the store on save must be repopulated into a freshly loaded
// Config so it decrypts back to its real value.
func TestLoad_RepopulatesSchemaRegistryCredentialFromStore(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	c.Contexts["orig"].SchemaRegistryClusters = map[string]*SchemaRegistryCluster{
		"lsrc-1": {
			Id:                     "lsrc-1",
			SchemaRegistryEndpoint: "https://sr.example.com",
			SrCredentials:          &APIKeyPair{Key: "SR-KEY", Secret: "sr-secret"},
		},
	}
	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())

	pair := reloaded.Contexts["orig"].SchemaRegistryClusters["lsrc-1"].SrCredentials
	require.NoError(t, pair.DecryptSecret())
	require.Equal(t, "sr-secret", pair.Secret)
}

// TestSave_SavedPasswordsKeyedByContextNotIdentity pins blocker 3: SavedCredentials is
// endpoint-specific (keyed by context name in memory), so two contexts that share one credential
// identity (same username, different URL/CA-cert) must each retain their own saved password across
// save/load. Keying the password by credential identity collapsed both into one record, so after
// reload both got the last-writer password and one login would fail.
func TestSave_SavedPasswordsKeyedByContextNotIdentity(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := New()
	c.Filename = filepath.Join(t.TempDir(), "config.json")

	// Two contexts sharing one credential identity ("shared-user"), different endpoints.
	c.Credentials["shared-user"] = &Credential{Name: "shared-user", Username: "shared-user", CredentialType: Username}
	addContext := func(name, server string) {
		c.Platforms[name] = &Platform{Name: name, Server: server}
		state := new(ContextState)
		ctx := &Context{
			Name:           name,
			PlatformName:   name,
			CredentialName: "shared-user",
			Platform:       c.Platforms[name],
			Credential:     c.Credentials["shared-user"],
			State:          state,
			Config:         c,
		}
		ctx.KafkaClusterContext = &KafkaClusterContext{KafkaClusterConfigs: map[string]*KafkaClusterConfig{}, Context: ctx}
		c.Contexts[name] = ctx
		c.ContextStates[name] = state
	}
	addContext("ctx-a", "https://a.example.com")
	addContext("ctx-b", "https://b.example.com")

	savePassword := func(ctxName, password string) {
		salt, nonce, err := secret.GenerateSaltAndNonce()
		require.NoError(t, err)
		encrypted, err := secret.Encrypt("shared-user", password, salt, nonce)
		require.NoError(t, err)
		c.SavedCredentials[ctxName] = &LoginCredential{Username: "shared-user", EncryptedPassword: encrypted, Salt: salt, Nonce: nonce}
	}
	savePassword("ctx-a", "password-a")
	savePassword("ctx-b", "password-b")

	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.GetFilename()
	require.NoError(t, reloaded.Load())

	decrypt := func(ctxName string) string {
		saved := reloaded.SavedCredentials[ctxName]
		require.NotNil(t, saved)
		plain, err := secret.Decrypt("shared-user", saved.EncryptedPassword, saved.Salt, saved.Nonce)
		require.NoError(t, err)
		return plain
	}
	require.Equal(t, "password-a", decrypt("ctx-a"), "ctx-a must keep its own password")
	require.Equal(t, "password-b", decrypt("ctx-b"), "ctx-b's password must not collide with ctx-a's shared identity")
}

func TestSecretsFilename_UnderStateDir(t *testing.T) {
	setTestHome(t, t.TempDir())
	require.Equal(t, stateDirPath("secrets.json"), SecretsFilename())
}

func TestIdentityKey_StableAcrossRename(t *testing.T) {
	c := newTestConfigWithAPIKeyContext(t)
	before := c.Contexts["orig"].identityKey()

	require.NoError(t, c.renameContextForTest("orig", "renamed"))

	require.Equal(t, before, c.Contexts["renamed"].identityKey())
}

// newTestConfigWithAPIKeyContext builds a valid Config with one api-key context named "orig",
// whose credential is "api-key-AK" (key "AK", secret "the-api-secret").
func newTestConfigWithAPIKeyContext(t *testing.T) *Config {
	t.Helper()
	c := New()
	c.Filename = filepath.Join(t.TempDir(), "config.json")
	require.NoError(t, c.Load())
	require.NoError(t, c.CreateContext("orig", "https://example.com", "AK", "the-api-secret"))
	return c
}

// renameContextForTest renames a context in place, moving its map entry and updating
// Context.Name while leaving CredentialName - the rename-stable identity key - untouched.
func (c *Config) renameContextForTest(oldName, newName string) error {
	ctx, ok := c.Contexts[oldName]
	if !ok {
		return fmt.Errorf(errors.ContextDoesNotExistErrorMsg, oldName)
	}
	if _, ok := c.Contexts[newName]; ok {
		return fmt.Errorf(errors.ContextAlreadyExistsErrorMsg, newName)
	}

	delete(c.Contexts, oldName)
	ctx.Name = newName
	c.Contexts[newName] = ctx

	if state, ok := c.ContextStates[oldName]; ok {
		delete(c.ContextStates, oldName)
		c.ContextStates[newName] = state
	}
	if c.CurrentContext == oldName {
		c.CurrentContext = newName
	}
	return nil
}
