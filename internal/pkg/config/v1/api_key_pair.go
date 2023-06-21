package v1

import (
	"runtime"
	"strings"

	"github.com/confluentinc/cli/internal/pkg/secret"
)

// APIKeyPair holds an API Key and Secret.
type APIKeyPair struct {
	Key    string `json:"api_key,omitempty"`
	Secret string `json:"api_secret,omitempty"`
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

func (c *APIKeyPair) EncryptSecret() error {
	if c.Salt == nil || c.Nonce == nil {
		salt, nonce, err := secret.GenerateSaltAndNonce()
		if err != nil {
			return err
		}
		c.Salt = salt
		c.Nonce = nonce
	}

	if !strings.HasPrefix(c.Secret, secret.AesGcm) {
		encryptedSecret, err := secret.Encrypt(c.Key, c.Secret, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.Secret = encryptedSecret
	}
	return nil
}
