package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/secret"
)

// each goroutine adds a different platform concurrently; before the lock and merge
// work, the last writer's whole-file overwrite dropped the others.
func TestSave_ConcurrentDifferentFields_NoLostWrite(t *testing.T) {
	setTestHome(t, t.TempDir())
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
func newSavedConfig(t *testing.T, path string) {
	t.Helper()

	c := New()
	c.Filename = path
	c.Platforms["platform"] = &Platform{Name: "platform", Server: "https://example.com"}
	c.Credentials["cred"] = &Credential{
		Name:           "cred",
		CredentialType: APIKey,
		APIKeyPair:     &APIKeyPair{Key: "api-key", Secret: "secret-original"},
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

// decryptToMatch must bring the baseline into the live config's full representation,
// not just its ciphertext: a secret's salt and nonce are part of its on-disk identity.
// If the baseline keeps a stale salt after its secret is decrypted, the three-way merge
// pairs disk's ciphertext with the wrong salt and decryption later fails.
func TestDecryptToMatch_AlignsSaltAndNonceWithDecryptedSecret(t *testing.T) {
	stored := &APIKeyPair{Key: "k", Secret: "plaintext"}
	require.NoError(t, stored.EncryptSecret()) // sets Salt/Nonce and encrypts

	base := New()
	base.Credentials["cred"] = &Credential{Name: "cred", APIKeyPair: &APIKeyPair{
		Key: "k", Secret: stored.Secret, Salt: stored.Salt, Nonce: stored.Nonce,
	}}
	// ref is a fresh plaintext pair with no salt/nonce, like a just-created credential.
	ref := New()
	ref.Credentials["cred"] = &Credential{Name: "cred", APIKeyPair: &APIKeyPair{Key: "k", Secret: "plaintext"}}

	require.NoError(t, base.decryptToMatch(ref))

	require.Equal(t, "plaintext", base.Credentials["cred"].APIKeyPair.Secret)
	require.Nil(t, base.Credentials["cred"].APIKeyPair.Salt,
		"baseline salt must match ref's (nil) so the merge keeps disk's ciphertext and salt together")
	require.Nil(t, base.Credentials["cred"].APIKeyPair.Nonce,
		"baseline nonce must match ref's (nil) so the merge keeps disk's ciphertext and nonce together")
}

// A context state's single salt/nonce is shared by both auth tokens. decryptToMatch must
// realign it to ref once it decrypts a token, but only when no token in the baseline is
// still encrypted: clearing the salt while the refresh token stays ciphertext would
// strand that ciphertext (its own load could no longer decrypt it).
func TestDecryptToMatch_ContextStateSaltAlignmentRespectsEncryptedToken(t *testing.T) {
	// Both tokens decrypted in ref: the baseline's salt/nonce realign to ref's (nil here).
	enc := encryptedState(t, "header.payload.signature", "v1.refreshtoken")

	base := New()
	base.ContextStates["ctx"] = &ContextState{AuthToken: enc.AuthToken, Salt: enc.Salt, Nonce: enc.Nonce}
	ref := New()
	ref.ContextStates["ctx"] = &ContextState{AuthToken: "header.payload.signature"} // plaintext, no salt

	require.NoError(t, base.decryptToMatch(ref))
	require.Nil(t, base.ContextStates["ctx"].Salt, "with both tokens plaintext, salt must realign to ref's nil")

	// Refresh token stays encrypted in ref (e.g. a Confluent Platform refresh token PreRun
	// leaves encrypted), so the baseline's salt must be preserved, not cleared.
	base2 := New()
	base2.ContextStates["ctx"] = &ContextState{
		AuthToken: enc.AuthToken, AuthRefreshToken: enc.AuthRefreshToken, Salt: enc.Salt, Nonce: enc.Nonce,
	}
	ref2 := New()
	ref2.ContextStates["ctx"] = &ContextState{
		AuthToken:        "header.payload.signature", // decrypted
		AuthRefreshToken: enc.AuthRefreshToken,       // still encrypted, shares the salt
		Salt:             enc.Salt,
		Nonce:            enc.Nonce,
	}

	require.NoError(t, base2.decryptToMatch(ref2))
	require.Equal(t, enc.Salt, base2.ContextStates["ctx"].Salt,
		"salt must be preserved while the refresh token stays encrypted")
}

// encryptedState returns a context state whose auth tokens are encrypted with a generated
// salt/nonce, produced by the real encryptSecrets path so the ciphertext, salt, and nonce
// are internally consistent.
func encryptedState(t *testing.T, authToken, authRefreshToken string) *ContextState {
	t.Helper()
	c := New()
	state := &ContextState{AuthToken: authToken, AuthRefreshToken: authRefreshToken}
	ctx := &Context{Name: "ctx", PlatformName: "https://confluent.cloud", State: state, Config: c}
	ctx.KafkaClusterContext = &KafkaClusterContext{Context: ctx}
	c.Contexts["ctx"] = ctx
	c.ContextStates["ctx"] = state
	c.CurrentContext = "ctx"

	require.NoError(t, c.encryptSecrets())
	require.True(t, isEncryptedSecret(state.AuthToken))
	require.True(t, isEncryptedSecret(state.AuthRefreshToken))
	return state
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
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path)

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
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path)

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
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path)

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
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path)

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

// On a fresh machine two sessions can both read the config as missing before either
// writes it. If one then creates it, the other's initial in-Load save must merge that
// file rather than overwrite it with a default. The sidecar lock is held here to force
// the exact window: the loading session reads the file missing, then blocks on the lock
// inside its save while another session's config lands on disk.
func TestLoad_FreshMachineDoesNotClobberConcurrentlyCreatedConfig(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	// Pause b right after it reads the file as missing, before its initial save, so the
	// competing config appears in exactly that window (deterministic, no timing race).
	reached := make(chan struct{})
	proceed := make(chan struct{})
	afterMissingConfigRead = func() {
		afterMissingConfigRead = func() {} // one-shot: b's own later saves are not gated
		close(reached)
		<-proceed
	}
	t.Cleanup(func() { afterMissingConfigRead = func() {} })

	b := New()
	b.Filename = path
	loadErr := make(chan error, 1)
	go func() { loadErr <- b.Load() }()

	<-reached // b has read the file as missing and is paused before its save

	other := New()
	other.Filename = path
	other.Platforms["from-other"] = &Platform{Name: "from-other", Server: "https://other.example.com"}
	data, err := json.MarshalIndent(other, "", "  ")
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, data, 0600))

	close(proceed) // b resumes into its locked save, which must merge rather than clobber
	require.NoError(t, <-loadErr)

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	require.Contains(t, final.Platforms, "from-other",
		"a config another session created after this session read the file missing must survive the fresh-machine load")
}

// addContextWithToken adds a second, non-current context carrying an auth token, shaped
// like newSavedConfig's context so Validate accepts it.
func addContextWithToken(c *Config, name, authToken string) {
	state := &ContextState{AuthToken: authToken}
	ctx := &Context{
		Name:           name,
		PlatformName:   "platform",
		CredentialName: "cred",
		Platform:       c.Platforms["platform"],
		Credential:     c.Credentials["cred"],
		State:          state,
		Config:         c,
	}
	ctx.KafkaClusterContext = &KafkaClusterContext{
		ActiveKafkaCluster:  "lkc-1",
		KafkaClusterConfigs: map[string]*KafkaClusterConfig{"lkc-1": {ID: "lkc-1", Name: "one"}},
		Context:             ctx,
	}
	c.Contexts[name] = ctx
	c.ContextStates[name] = state
}

// Contexts and ContextStates are two halves of one persisted unit. When one session
// deletes a context while this process edits that same context, an independent per-map
// merge keeps the edited Context but drops its unedited state, orphaning it so Validate
// drops the auth token or rejects the save. The two halves must stay coupled.
func TestSave_ConcurrentContextDeleteVsEdit_KeepsStateWithContext(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path) // context "ctx" (current)

	seed := loadDecrypted(t, path)
	addContextWithToken(seed, "target", "org-1-target.jwt.token") // a second, non-current context
	require.NoError(t, seed.Save())

	a := loadDecrypted(t, path) // this process
	b := loadDecrypted(t, path) // concurrent session

	require.NoError(t, b.DeleteContext("target")) // concurrent delete of both halves

	a.Contexts["target"].CurrentEnvironment = "env-z" // this process edits the same context
	require.NoError(t, a.Save())

	final := loadDecrypted(t, path)
	require.Contains(t, final.Contexts, "target", "the edited context must survive")
	require.Contains(t, final.ContextStates, "target", "its state must survive with it, not be orphaned")
	require.NotEmpty(t, final.ContextStates["target"].AuthToken,
		"the context's auth token must not be dropped by the concurrent delete")
}

// newTwoIdentityConfig seeds two independent identities (own credential, own context, own
// token) under path, so a concurrent token rewrite to one must not clobber the other.
func newTwoIdentityConfig(t *testing.T, path string) {
	t.Helper()

	c := New()
	c.Filename = path
	c.Platforms["platform"] = &Platform{Name: "platform", Server: "https://example.com"}

	addIdentity := func(name string) {
		credName := "cred-" + name
		c.Credentials[credName] = &Credential{
			Name:           credName,
			CredentialType: APIKey,
			APIKeyPair:     &APIKeyPair{Key: "api-key-" + name, Secret: "secret-" + name},
		}
		state := &ContextState{AuthToken: "header.payload." + name}
		ctx := &Context{
			Name:           name,
			PlatformName:   "platform",
			CredentialName: credName,
			Platform:       c.Platforms["platform"],
			Credential:     c.Credentials[credName],
			State:          state,
			Config:         c,
		}
		ctx.KafkaClusterContext = &KafkaClusterContext{
			ActiveKafkaCluster:  "lkc-1",
			KafkaClusterConfigs: map[string]*KafkaClusterConfig{"lkc-1": {ID: "lkc-1", Name: "one"}},
			Context:             ctx,
		}
		c.Contexts[name] = ctx
		c.ContextStates[name] = state
	}
	addIdentity("a")
	addIdentity("b")
	c.CurrentContext = "a"

	require.NoError(t, c.Save())
}

// Two sessions each rewrite a DIFFERENT identity's token, interleaved. Before the secret
// store's own three-way merge, saveSecretStore rebuilt and overwrote the whole record set
// from whichever process saved last, so the earlier session's rotation of its own identity's
// token was lost the moment the other session (which never touched it) saved afterward.
func TestSecretStore_ConcurrentDifferentIdentitiesBothPersist(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newTwoIdentityConfig(t, path)

	a := loadDecrypted(t, path)
	b := loadDecrypted(t, path)

	a.ContextStates["a"].AuthToken = "header.payload.a-rotated"
	require.NoError(t, a.Save())

	b.ContextStates["b"].AuthToken = "header.payload.b-rotated"
	require.NoError(t, b.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	require.NoError(t, final.ContextStates["a"].DecryptAuthToken("a"))
	require.NoError(t, final.ContextStates["b"].DecryptAuthToken("b"))
	require.Equal(t, "header.payload.a-rotated", final.ContextStates["a"].AuthToken,
		"a's own token rotation must survive b's concurrent, unrelated save")
	require.Equal(t, "header.payload.b-rotated", final.ContextStates["b"].AuthToken,
		"b's own token rotation must survive too - both identities persist")
}

// A concurrent logout (cleared token) must not be resurrected by an unrelated save from a
// session whose baseline still holds the old token. Without a baseline to diff against,
// saveSecretStore cannot tell "b never touched this token" from "b's stale copy should win",
// so it would re-encrypt b's stale plaintext and bring the cleared token back from the dead.
func TestSecretStore_ConcurrentLogoutNotResurrected(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path) // context "ctx" (current), no token yet

	seed := loadDecrypted(t, path)
	seed.ContextStates["ctx"].AuthToken = "header.payload.original"
	require.NoError(t, seed.Save())

	a := loadDecrypted(t, path) // will log out
	b := loadDecrypted(t, path) // stale baseline, saves something unrelated afterward

	a.ContextStates["ctx"].AuthToken = ""
	a.ContextStates["ctx"].AuthRefreshToken = ""
	require.NoError(t, a.Save())

	b.Contexts["ctx"].CurrentEnvironment = "env-from-b"
	require.NoError(t, b.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	require.Empty(t, final.ContextStates["ctx"].AuthToken,
		"a's concurrent logout must not be resurrected by b's unrelated, stale-baseline save")
	require.Equal(t, "env-from-b", final.Contexts["ctx"].CurrentEnvironment)
}

// tamperTokenCiphertext overwrites cfg's OWN in-memory secret baseline for ctxName with a
// different encryption of the SAME plaintext (fresh salt/nonce). This reproduces, on any
// platform, the exact shape Windows produces on every save: DPAPI is non-deterministic and
// GenerateSaltAndNonce returns nil,nil there, so re-encrypting an untouched plaintext token
// yields different ciphertext bytes than what secretBaseline holds, even though nothing
// about the token actually changed.
func tamperTokenCiphertext(t *testing.T, cfg *Config, ctxName, plaintext string) {
	t.Helper()
	altSalt, altNonce, err := secret.GenerateSaltAndNonce()
	require.NoError(t, err)
	altCiphertext, err := secret.Encrypt(ctxName, plaintext, altSalt, altNonce)
	require.NoError(t, err)
	cfg.secretBaseline.Tokens[ctxName] = &tokenRecord{AuthToken: altCiphertext, Salt: altSalt, Nonce: altNonce}
}

// A ciphertext-only diff (as if re-encrypting an untouched plaintext always reproduced the
// same bytes, true on Unix but NOT on Windows: DPAPI is non-deterministic and
// GenerateSaltAndNonce returns nil,nil there) would see a's own untouched token as "changed"
// the moment its baseline's ciphertext differs by even one byte from a fresh re-encryption -
// even though the plaintext a holds is identical to what the baseline decrypts to. This
// reproduces that exact shape on Unix (see tamperTokenCiphertext) and pins that the merge
// must decide on plaintext: a's tampered baseline must not make its own untouched token look
// locally changed and clobber a concurrent rotation it never touched.
func TestSecretStore_UnchangedPlaintextDifferentCiphertext_NotClobbered(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path) // context "ctx" (current), no token yet

	seed := loadDecrypted(t, path)
	seed.ContextStates["ctx"].AuthToken = "header.payload.original"
	require.NoError(t, seed.Save())

	a := loadDecrypted(t, path) // holds "header.payload.original" plaintext live, untouched
	tamperTokenCiphertext(t, a, "ctx", "header.payload.original")

	other := loadDecrypted(t, path) // concurrent session, rotates the token
	other.ContextStates["ctx"].AuthToken = "header.payload.rotated"
	require.NoError(t, other.Save())

	a.Contexts["ctx"].CurrentEnvironment = "env-from-a" // a's own, unrelated edit
	require.NoError(t, a.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	require.NoError(t, final.ContextStates["ctx"].DecryptAuthToken("ctx"))
	require.Equal(t, "header.payload.rotated", final.ContextStates["ctx"].AuthToken,
		"a's tampered baseline ciphertext must not make an untouched token look locally changed and clobber the concurrent rotation")
}

// The same tampered-ciphertext hazard on the logout-not-resurrected guarantee: a's baseline
// disagrees with a fresh re-encryption of the SAME plaintext (see tamperTokenCiphertext), but
// a never touched the token. A ciphertext-only diff would misread that as a local change and
// resurrect the concurrent logout; deciding on plaintext must not.
func TestSecretStore_UnchangedPlaintextDifferentCiphertext_LogoutStaysCleared(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newSavedConfig(t, path)

	seed := loadDecrypted(t, path)
	seed.ContextStates["ctx"].AuthToken = "header.payload.original"
	require.NoError(t, seed.Save())

	a := loadDecrypted(t, path)
	tamperTokenCiphertext(t, a, "ctx", "header.payload.original")

	other := loadDecrypted(t, path) // concurrent session, logs out
	other.ContextStates["ctx"].AuthToken = ""
	other.ContextStates["ctx"].AuthRefreshToken = ""
	require.NoError(t, other.Save())

	a.Contexts["ctx"].CurrentEnvironment = "env-from-a" // a's own, unrelated edit
	require.NoError(t, a.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	require.Empty(t, final.ContextStates["ctx"].AuthToken,
		"a's tampered baseline ciphertext must not make an untouched token look locally changed and resurrect the concurrent logout")
}

// newConfigWithNestedKeys seeds a config with one context holding a Global API key and a
// cluster-scoped Kafka API key (both with secrets) under path.
func newConfigWithNestedKeys(t *testing.T, path string) {
	t.Helper()

	c := New()
	c.Filename = path
	c.Platforms["platform"] = &Platform{Name: "platform", Server: "https://example.com"}
	c.Credentials["cred"] = &Credential{
		Name:           "cred",
		CredentialType: APIKey,
		APIKeyPair:     &APIKeyPair{Key: "api-key", Secret: "secret-original"},
	}
	state := new(ContextState)
	ctx := &Context{
		Name:           "ctx",
		PlatformName:   "platform",
		CredentialName: "cred",
		Platform:       c.Platforms["platform"],
		Credential:     c.Credentials["cred"],
		State:          state,
		Config:         c,
		GlobalAPIKeys:  map[string]*APIKeyPair{},
	}
	cluster := &KafkaClusterConfig{
		ID:        "lkc-1",
		Name:      "one",
		Bootstrap: "https://example.com",
		APIKeys:   map[string]*APIKeyPair{"CK": {Key: "CK", Secret: "ck-original"}},
	}
	require.NoError(t, cluster.EncryptAPIKeys())
	ctx.KafkaClusterContext = &KafkaClusterContext{
		KafkaClusterConfigs: map[string]*KafkaClusterConfig{"lkc-1": cluster},
		Context:             ctx,
	}
	c.Contexts["ctx"] = ctx
	c.ContextStates["ctx"] = state
	c.CurrentContext = "ctx"

	require.NoError(t, ctx.StoreGlobalAPIKey(&APIKeyPair{Key: "GK", Secret: "gk-original"}))
	require.NoError(t, c.Save())
}

// tamperNestedKeyCiphertext overwrites cfg's OWN in-memory secret baseline for a nested API key
// (Global or, when clusterId != "", a Kafka cluster key) with a different encryption of the SAME
// plaintext (fresh salt/nonce). This reproduces, on any platform, the shape Windows produces on
// every save: ResolveKafkaAPIKey decrypts a nested key in place and it re-encrypts to different
// ciphertext, even though nothing about the key's plaintext changed.
func tamperNestedKeyCiphertext(t *testing.T, cfg *Config, identity, clusterId, keyId, plaintext string) {
	t.Helper()
	pair := &APIKeyPair{Key: keyId, Secret: plaintext}
	require.NoError(t, pair.EncryptSecret())
	triple := &apiKeySecret{Secret: pair.Secret, Salt: pair.Salt, Nonce: pair.Nonce}

	rec := cfg.secretBaseline.Secrets[identity]
	require.NotNil(t, rec)
	if clusterId == "" {
		rec.GlobalAPIKeys[keyId] = triple
	} else {
		rec.KafkaAPIKeys[clusterId][keyId] = triple
	}
}

// A ciphertext-only diff would see a's own untouched Global API key as "changed" the moment its
// baseline's ciphertext differs from a fresh re-encryption (true on Windows: ResolveKafkaAPIKey
// decrypts nested keys in place, which then re-encrypt non-deterministically). This reproduces that
// shape on Unix (tamperNestedKeyCiphertext) and pins that the nested-key merge decides on plaintext:
// a's churned baseline must not make its own untouched key look changed and clobber a concurrent
// rotation it never touched.
func TestSecretStore_NestedGlobalKeyUnchangedPlaintextDifferentCiphertext_NotClobbered(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newConfigWithNestedKeys(t, path)

	a := loadDecrypted(t, path) // holds "gk-original" (ciphertext) live, untouched
	tamperNestedKeyCiphertext(t, a, "cred", "", "GK", "gk-original")

	other := loadDecrypted(t, path) // concurrent session, rotates the Global key
	rotated := other.Contexts["ctx"].GlobalAPIKeys["GK"]
	rotated.Secret, rotated.Salt, rotated.Nonce = "gk-rotated", nil, nil
	require.NoError(t, other.Save())

	a.Contexts["ctx"].CurrentEnvironment = "env-from-a" // a's own, unrelated edit
	require.NoError(t, a.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	pair := final.Contexts["ctx"].GlobalAPIKeys["GK"]
	require.NoError(t, pair.DecryptSecret())
	require.Equal(t, "gk-rotated", pair.Secret,
		"a's churned baseline ciphertext must not make an untouched Global key look changed and clobber the concurrent rotation")
}

// The Kafka cluster-scoped analog of the Global-key churn test above.
func TestSecretStore_NestedKafkaKeyUnchangedPlaintextDifferentCiphertext_NotClobbered(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newConfigWithNestedKeys(t, path)

	a := loadDecrypted(t, path)
	tamperNestedKeyCiphertext(t, a, "cred", "lkc-1", "CK", "ck-original")

	other := loadDecrypted(t, path) // concurrent session, rotates the cluster key
	rotated := other.Contexts["ctx"].KafkaClusterContext.KafkaClusterConfigs["lkc-1"].APIKeys["CK"]
	rotated.Secret, rotated.Salt, rotated.Nonce = "ck-rotated", nil, nil
	require.NoError(t, other.Save())

	a.Contexts["ctx"].CurrentEnvironment = "env-from-a"
	require.NoError(t, a.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	pair := final.Contexts["ctx"].KafkaClusterContext.KafkaClusterConfigs["lkc-1"].APIKeys["CK"]
	require.NoError(t, pair.DecryptSecret())
	require.Equal(t, "ck-rotated", pair.Secret,
		"a's churned baseline ciphertext must not make an untouched Kafka key look changed and clobber the concurrent rotation")
}

// A nested API key added ONLY by a concurrent session arrives in this process's save via the
// config.json merge (disk -> merged) carrying just its public id; its secret is json:"-", so
// merged's copy reads empty and merged.Validate() would prune the id as "malformed" before the
// write. rehydrateNestedAPIKeySecretPresence must restore presence for such disk-only keys from the
// on-disk secret store too, not only for keys this process's own live config knows about, so the
// concurrently-added key survives (and its secret rides through the secret-store merge).
func TestSecretStore_ConcurrentDiskOnlyNestedKeyNotPruned(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	newConfigWithNestedKeys(t, path)

	a := loadDecrypted(t, path) // never learns about the concurrently-added key

	other := loadDecrypted(t, path)
	require.NoError(t, other.Contexts["ctx"].StoreGlobalAPIKey(&APIKeyPair{Key: "GK2", Secret: "gk2-secret"}))
	require.NoError(t, other.Save())

	a.Contexts["ctx"].CurrentEnvironment = "env-from-a" // a's own, unrelated edit
	require.NoError(t, a.Save())

	final := New()
	final.Filename = path
	require.NoError(t, final.Load())
	pair := final.Contexts["ctx"].GlobalAPIKeys["GK2"]
	require.NotNil(t, pair, "a concurrently-added nested key must not be pruned by this process's Validate")
	require.NoError(t, pair.DecryptSecret())
	require.Equal(t, "gk2-secret", pair.Secret)
}

// save() must resolve flag overrides exactly once. It is only reached from saveLocked
// (via writeWholeConfig), which already swapped flag values out for the persisted ones;
// a second resolve here re-applies them against the now-switched current context and
// writes one context's --cluster value onto another. Uses the constructed-config
// (nil baseline) whole-write path.
func TestSave_WholeConfigWrite_DoesNotDoubleResolveOverrides(t *testing.T) {
	setTestHome(t, t.TempDir())
	c := New()
	c.Filename = filepath.Join(t.TempDir(), "config.json")
	c.Platforms["platform"] = &Platform{Name: "platform", Server: "https://example.com"}
	c.Credentials["cred"] = &Credential{Name: "cred", CredentialType: APIKey, APIKeyPair: &APIKeyPair{Key: "k", Secret: "s"}}

	addCtx := func(name, activeKafka string) {
		state := new(ContextState)
		ctx := &Context{
			Name: name, PlatformName: "platform", CredentialName: "cred",
			Platform: c.Platforms["platform"], Credential: c.Credentials["cred"], State: state, Config: c,
		}
		ctx.KafkaClusterContext = &KafkaClusterContext{
			ActiveKafkaCluster: activeKafka,
			KafkaClusterConfigs: map[string]*KafkaClusterConfig{
				"kA": {ID: "kA", Name: "a"}, "kB": {ID: "kB", Name: "b"}, "kX": {ID: "kX", Name: "x"},
			},
			Context: ctx,
		}
		c.Contexts[name] = ctx
		c.ContextStates[name] = state
	}
	addCtx("A", "kA")
	addCtx("B", "kX") // B's active currently holds the --cluster flag value

	// Simulate `--context B --cluster kX`: current switched to B, real values stashed.
	c.CurrentContext = "B"
	c.overwrittenCurrentContext = "A"       // A is the real current context
	c.overwrittenCurrentKafkaCluster = "kB" // kB is B's real active cluster

	require.NoError(t, c.Save())

	reloaded := New()
	reloaded.Filename = c.Filename
	require.NoError(t, reloaded.Load())
	require.Equal(t, "kA", reloaded.Contexts["A"].KafkaClusterContext.GetActiveKafkaClusterId(),
		"context A's cluster must not be overwritten by B's --cluster value through a second resolve")
	require.Equal(t, "A", reloaded.CurrentContext, "the persisted current context must be the real one, not the --context flag")
}

// Two contexts created in separate load/save cycles that reuse one API key (so one
// credential is overwritten) must leave that credential decryptable. The second
// create loads the config (credential still encrypted, its salt present) and replaces
// the credential with a fresh plaintext pair that has no salt. The merge must keep the
// credential's ciphertext, salt, and nonce as one unit; if decryptToMatch aligns only
// the ciphertext, the merge pairs disk's ciphertext with a regenerated salt and the
// next load fails GCM authentication. Reproduces the integration-test flow in
// test/context_test.go's contextCreateArgs.
func TestSave_TwoContextCreatesSharedCredential_SecretStaysDecryptable(t *testing.T) {
	setTestHome(t, t.TempDir())
	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")

	createContextReusingAPIKey(t, path, "0")
	createContextReusingAPIKey(t, path, "1")

	next := New()
	next.Filename = path
	require.NoError(t, next.Load())
	require.NoError(t, next.DecryptCredentials(),
		"a credential overwritten across two create cycles must stay decryptable")
	require.Equal(t, "api-secret-value", next.Credentials["api-key-test"].APIKeyPair.Secret)
}

// createContextReusingAPIKey mirrors test/context_test.go's contextCreateArgs: a fresh
// New()+Load()+CreateContext per call, all using the same API key so the derived
// credential name collides and the credential is overwritten on the second call.
func createContextReusingAPIKey(t *testing.T, path, name string) {
	t.Helper()
	cfg := New()
	cfg.Filename = path
	require.NoError(t, cfg.Load())
	require.NoError(t, cfg.CreateContext(name, "https://example.com", "test", "api-secret-value"))
}
