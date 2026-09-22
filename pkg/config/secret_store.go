package config

// secretRecord holds the decrypted-at-rest secret material for one credential identity.
type secretRecord struct {
	AuthToken        string `json:"auth_token,omitempty"`
	AuthRefreshToken string `json:"auth_refresh_token,omitempty"`
	Secret           string `json:"secret,omitempty"`   // api-key secret
	Password         string `json:"password,omitempty"` // saved login password
	Salt             []byte `json:"salt,omitempty"`
	Nonce            []byte `json:"nonce,omitempty"`
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

// identityKey is the rename-stable key for a context's secrets: CredentialName is set once at
// context creation and untouched by a later context rename.
func (c *Context) identityKey() string {
	return c.CredentialName
}
