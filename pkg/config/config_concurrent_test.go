package config

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"
)

// each goroutine adds a different platform concurrently; before the lock and merge
// work, the last writer's whole-file overwrite dropped the others.
func TestSave_ConcurrentDifferentFields_NoLostWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	seed := New()
	seed.Filename = path
	require.NoError(t, seed.Save())

	const n = 8
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c, err := readConfigFromDisk(path, seed)
			require.NoError(t, err)
			c.snapshotBaseline()
			name := fmt.Sprintf("platform-%d", i)
			c.Platforms[name] = &Platform{Name: name}
			require.NoError(t, c.Save())
		}(i)
	}
	wg.Wait()

	final, err := readConfigFromDisk(path, seed)
	require.NoError(t, err)
	for i := 0; i < n; i++ {
		require.Contains(t, final.Platforms, fmt.Sprintf("platform-%d", i),
			"every concurrent session's platform must survive")
	}
}

// helper: a saved single-context config with an API-key credential and two known
// Kafka clusters (so switching the active cluster passes Validate).
func newSavedConfig(t *testing.T, path, apiSecret string) {
	t.Helper()

	c := New()
	c.Filename = path
	c.Platforms["platform"] = &Platform{Name: "platform", Server: "https://example.com"}
	c.Credentials["cred"] = &Credential{
		Name:           "cred",
		CredentialType: APIKey,
		APIKeyPair:     &APIKeyPair{Key: "api-key", Secret: apiSecret},
	}
	state := new(ContextState)
	ctx := &Context{
		Name:               "ctx",
		PlatformName:       "platform",
		CredentialName:     "cred",
		CurrentEnvironment: "env-a",
		Platform:           c.Platforms["platform"],
		Credential:         c.Credentials["cred"],
		State:              state,
		Config:             c,
	}
	ctx.KafkaClusterContext = &KafkaClusterContext{
		ActiveKafkaCluster: "lkc-1",
		KafkaClusterConfigs: map[string]*KafkaClusterConfig{
			"lkc-1": {ID: "lkc-1", Name: "one"},
			"lkc-2": {ID: "lkc-2", Name: "two"},
		},
		Context: ctx,
	}
	c.Contexts["ctx"] = ctx
	c.ContextStates["ctx"] = state
	c.CurrentContext = "ctx"

	require.NoError(t, c.Save())
}

// helper: load and decrypt as PreRun does, leaving the live config plaintext while
// the merge baseline stays encrypted (the split the three-way merge must tolerate).
func loadDecrypted(t *testing.T, path string) *Config {
	t.Helper()

	c := New()
	c.Filename = path
	require.NoError(t, c.Load())
	require.NoError(t, c.DecryptCredentials())
	require.NoError(t, c.DecryptContextStates())
	return c
}

// decryptToMatch must align the encrypted baseline to the live plaintext so an
// untouched secret is not seen as changed. This holds by decryption, independent of
// whether re-encryption would reproduce the ciphertext (it does not on Windows/DPAPI).
func TestDecryptToMatch_AlignsBaselineToLivePlaintext(t *testing.T) {
	plaintext := "top-secret"
	stored := &APIKeyPair{Key: "k", Secret: plaintext}
	require.NoError(t, stored.EncryptSecret())
	require.True(t, isEncryptedSecret(stored.Secret))

	base := New()
	base.Credentials["cred"] = &Credential{Name: "cred", APIKeyPair: &APIKeyPair{
		Key: "k", Secret: stored.Secret, Salt: stored.Salt, Nonce: stored.Nonce,
	}}
	ref := New()
	ref.Credentials["cred"] = &Credential{Name: "cred", APIKeyPair: &APIKeyPair{Key: "k", Secret: plaintext}}

	require.NoError(t, base.decryptToMatch(ref))

	require.Equal(t, plaintext, base.Credentials["cred"].APIKeyPair.Secret,
		"baseline secret must be decrypted to the live plaintext, so an untouched secret is not seen as changed")
}

// A stored plaintext token (e.g. a cloud refresh token, which is never encrypted)
// carries no cipher prefix but may sit beside a salt. decryptToMatch must not feed
// it to Decrypt, which would fail authentication on a value that was never encrypted.
func TestDecryptToMatch_SkipsPlaintextBaselineTokens(t *testing.T) {
	base := New()
	base.ContextStates["ctx"] = &ContextState{
		AuthRefreshToken: "refreshToken", // plaintext, no cipher prefix
		Salt:             []byte("0123456789012345678901234"),
		Nonce:            []byte("012345678901"),
	}
	ref := New()
	ref.ContextStates["ctx"] = &ContextState{AuthRefreshToken: "refreshToken"}

	require.NoError(t, base.decryptToMatch(ref), "a plaintext baseline token must be left as-is, not decrypted")
	require.Equal(t, "refreshToken", base.ContextStates["ctx"].AuthRefreshToken)
}

// A merge can leave a plaintext token in a context that is not current at write time
// (e.g. after a concurrent `context use` switch). encryptSecrets must encrypt every
// context's token, not only the current one, so none reaches disk plaintext.
func TestEncryptSecrets_EncryptsNonCurrentContextTokens(t *testing.T) {
	c := New()
	for _, name := range []string{"A", "B"} {
		state := &ContextState{AuthToken: "org-1-y0ur.jwt.T0kEn"} // plaintext JWT-shaped token
		ctx := &Context{Name: name, PlatformName: "https://confluent.cloud", State: state, Config: c}
		ctx.KafkaClusterContext = &KafkaClusterContext{Context: ctx}
		c.Contexts[name] = ctx
		c.ContextStates[name] = state
	}
	c.CurrentContext = "B" // A is not current

	require.NoError(t, c.encryptSecrets())

	require.True(t, isEncryptedSecret(c.ContextStates["A"].AuthToken), "a non-current context's plaintext token must be encrypted")
	require.True(t, isEncryptedSecret(c.ContextStates["B"].AuthToken), "the current context's token must be encrypted")
}

// Two sessions change different fields of the SAME context (environment vs. active
// Kafka cluster). Both live inside one Contexts map value, so an atomic per-key
// overlay drops the earlier writer's field; the merge must combine them.
func TestSave_ConcurrentSameContextDifferentFields_NoLostWrite(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path, "secret-original")

	a := loadDecrypted(t, path)
	b := loadDecrypted(t, path)

	a.Contexts["ctx"].CurrentEnvironment = "env-b" // like `confluent environment use`
	require.NoError(t, a.Save())

	b.Contexts["ctx"].KafkaClusterContext.SetActiveKafkaCluster("lkc-2") // like `confluent kafka cluster use`
	require.NoError(t, b.Save())

	final := loadDecrypted(t, path)
	require.Equal(t, "env-b", final.Contexts["ctx"].CurrentEnvironment,
		"the environment change must survive the concurrent Kafka-cluster save")
	require.Equal(t, "lkc-2", final.Contexts["ctx"].KafkaClusterContext.GetActiveKafkaClusterId(),
		"the concurrent Kafka-cluster change must survive")
}

// A session that never touched a shared credential must not clobber another
// session's concurrent update to it. The live copy is decrypted while the baseline
// is encrypted, so a representation-blind diff wrongly reads the credential as
// locally changed and overwrites the concurrent rotation.
func TestSave_ConcurrentSharedCredential_NotClobbered(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path, "secret-original")

	b := loadDecrypted(t, path) // holds decrypted "secret-original"
	a := loadDecrypted(t, path)

	a.Credentials["cred"].APIKeyPair.Secret = "secret-rotated"
	require.NoError(t, a.Save())

	// b never touched the credential; it changes an unrelated field and saves after a.
	b.Contexts["ctx"].CurrentEnvironment = "env-from-b"
	require.NoError(t, b.Save())

	final := loadDecrypted(t, path)
	require.Equal(t, "secret-rotated", final.Credentials["cred"].APIKeyPair.Secret,
		"a's concurrent credential rotation must not be clobbered by b's unrelated save")
	require.Equal(t, "env-from-b", final.Contexts["ctx"].CurrentEnvironment)
}

// A process that merges a concurrent disk change into its first save must not revert
// that change on a later save in the same process. The merge writes the concurrent
// field to disk but leaves the live config's copy stale, so if the post-save baseline
// tracks the merged (disk) value while ours stays stale, the next diff misreads the
// untouched field as a local edit and overwrites the concurrent change.
func TestSave_SecondSaveInSameProcess_PreservesConcurrentDiskChange(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path, "secret-original")

	p := loadDecrypted(t, path)     // this process, baseline captured at load
	other := loadDecrypted(t, path) // a concurrent session
	require.True(t, p.EnableColor, "precondition: color starts enabled (New()'s default)")

	other.EnableColor = false // a field p never touches
	require.NoError(t, other.Save())

	p.Contexts["ctx"].CurrentEnvironment = "env-b" // p's first, unrelated edit
	require.NoError(t, p.Save())

	mid := loadDecrypted(t, path)
	require.False(t, mid.EnableColor, "the first save must merge in the concurrent color change")

	p.Contexts["ctx"].CurrentEnvironment = "env-c" // p's second edit, still not touching color
	require.NoError(t, p.Save())

	final := loadDecrypted(t, path)
	require.False(t, final.EnableColor, "the second save must not revert the concurrently-disabled color")
	require.Equal(t, "env-c", final.Contexts["ctx"].CurrentEnvironment)
}

// The same second-save hazard on a secret rather than a scalar. The post-save baseline
// is a snapshot of the live config, whose credential secret is plaintext, while disk
// holds ciphertext; this pins that decryptToMatch still aligns the two so a concurrent
// rotation this process never touched survives a later save in the same process.
func TestSave_SecondSaveInSameProcess_PreservesConcurrentSecretRotation(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path, "secret-original")

	p := loadDecrypted(t, path)     // this process, holds decrypted "secret-original"
	other := loadDecrypted(t, path) // a concurrent session

	other.Credentials["cred"].APIKeyPair.Secret = "secret-rotated" // a field p never touches
	require.NoError(t, other.Save())

	p.Contexts["ctx"].CurrentEnvironment = "env-b" // p's first, unrelated edit
	require.NoError(t, p.Save())

	mid := loadDecrypted(t, path)
	require.Equal(t, "secret-rotated", mid.Credentials["cred"].APIKeyPair.Secret,
		"the first save must merge in the concurrent rotation")

	p.Contexts["ctx"].CurrentEnvironment = "env-c" // p's second edit, still not touching the secret
	require.NoError(t, p.Save())

	final := loadDecrypted(t, path)
	require.Equal(t, "secret-rotated", final.Credentials["cred"].APIKeyPair.Secret,
		"the second save must not revert the concurrent rotation")
	require.Equal(t, "env-c", final.Contexts["ctx"].CurrentEnvironment)
}
