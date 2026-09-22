package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

// secretRecord holds one credential identity's secret material as stored on disk: each
// value is ciphertext paired with its own salt/nonce. A login identity can hold auth
// tokens AND a saved password, encrypted independently, so each secret kind carries its
// own salt/nonce pair rather than sharing one - pairing a ciphertext with the wrong
// salt/nonce fails GCM authentication on load.
type secretRecord struct {
	// AuthToken and AuthRefreshToken share one salt/nonce, from ContextState.
	AuthToken        string `json:"auth_token,omitempty"`
	AuthRefreshToken string `json:"auth_refresh_token,omitempty"`
	TokenSalt        []byte `json:"token_salt,omitempty"`
	TokenNonce       []byte `json:"token_nonce,omitempty"`

	// Secret is the api-key secret, from APIKeyPair.
	Secret      string `json:"secret,omitempty"`
	SecretSalt  []byte `json:"secret_salt,omitempty"`
	SecretNonce []byte `json:"secret_nonce,omitempty"`

	// Password is the saved login password, from LoginCredential.
	Password      string `json:"password,omitempty"`
	PasswordSalt  []byte `json:"password_salt,omitempty"`
	PasswordNonce []byte `json:"password_nonce,omitempty"`
}

// secretFile is the on-disk shape of the secret store, keyed by credential identity
// (Context.identityKey) so contexts sharing a login share one entry and a context rename
// never orphans it.
type secretFile struct {
	Secrets map[string]*secretRecord `json:"secrets,omitempty"`
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

// saveSecretStore extracts the secret material already encrypted in c's contexts,
// credentials, and saved passwords into the secret store, keyed by identity. Callers
// must run it after encryptSecrets, under the Save lock, so it captures ciphertext
// rather than the plaintext the live config holds mid-session.
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

	for _, ctx := range c.Contexts {
		state := ctx.GetState()
		if state == nil || (state.AuthToken == "" && state.AuthRefreshToken == "") {
			continue
		}
		r := record(ctx.identityKey())
		r.AuthToken = state.AuthToken
		r.AuthRefreshToken = state.AuthRefreshToken
		r.TokenSalt = state.Salt
		r.TokenNonce = state.Nonce
	}

	for name, credential := range c.Credentials {
		if credential == nil || credential.APIKeyPair == nil || credential.APIKeyPair.Secret == "" {
			continue
		}
		r := record(name)
		r.Secret = credential.APIKeyPair.Secret
		r.SecretSalt = credential.APIKeyPair.Salt
		r.SecretNonce = credential.APIKeyPair.Nonce
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

	if len(records) == 0 {
		return nil
	}

	return newSecretStore().write(&secretFile{Secrets: records})
}
