package config

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/errors"
)

// TestSave_SecretsLeaveConfigFile pins the extraction contract: a saved API-key secret
// must not appear in config.json (plaintext or ciphertext) and must land, still
// encrypted, in the secret store.
func TestSave_SecretsLeaveConfigFile(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := newTestConfigWithAPIKeyContext(t)
	require.NoError(t, c.Save())

	cfgRaw, err := os.ReadFile(c.GetFilename())
	require.NoError(t, err)
	require.NotContains(t, string(cfgRaw), "the-api-secret")
	require.NotContains(t, string(cfgRaw), `"secret"`)

	secRaw, err := os.ReadFile(SecretsFilename())
	require.NoError(t, err)
	require.Contains(t, string(secRaw), "secrets")
	require.NotContains(t, string(secRaw), "the-api-secret") // encrypted, not plaintext
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
