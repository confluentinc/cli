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

	plainAuthToken, err := secret.Decrypt("orig", rec.AuthToken, rec.TokenSalt, rec.TokenNonce)
	require.NoError(t, err)
	require.Equal(t, "header.payload.signature", plainAuthToken)

	plainRefreshToken, err := secret.Decrypt("orig", rec.AuthRefreshToken, rec.TokenSalt, rec.TokenNonce)
	require.NoError(t, err)
	require.Equal(t, "v1.some-refresh-token", plainRefreshToken)

	plainPassword, err := secret.Decrypt("orig-user", rec.Password, rec.PasswordSalt, rec.PasswordNonce)
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
