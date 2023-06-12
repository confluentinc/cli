package v1

import (
	"runtime"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/secret"
)

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key"`
	Secret string `json:"api_secret"`
	Salt   []byte `json:"salt,omitempty"`
	Nonce  []byte `json:"nonce,omitempty"`
}

func (c *APIKeyPair) DecryptSecret() error {
	if c.Secret != "" && strings.HasPrefix(c.Secret, secret.AesGcm) && (c.Salt != nil || runtime.GOOS == "windows") {
		decryptedSecret, err := secret.Decrypt(c.Key, c.Secret, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.Secret = decryptedSecret
	}

	return nil
}
