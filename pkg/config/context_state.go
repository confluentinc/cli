package config

import (
	"runtime"
	"strings"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"

	"github.com/confluentinc/cli/v3/pkg/secret"
)

type ContextState struct {
	// Deprecated
	Auth *AuthConfig `json:"auth,omitempty"`

	AuthToken        string `json:"auth_token"`
	AuthRefreshToken string `json:"auth_refresh_token"`
	Salt             []byte `json:"salt,omitempty"`
	Nonce            []byte `json:"nonce,omitempty"`
}

func (c *ContextState) DecryptAuthToken(ctxName string) error {
	if (strings.HasPrefix(c.AuthToken, secret.AesGcm) && c.Salt != nil) || (c.AuthToken != "" && runtime.GOOS == "windows") {
		decryptedAuthToken, err := secret.Decrypt(ctxName, c.AuthToken, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.AuthToken = decryptedAuthToken
	}

	return nil
}

func (c *ContextState) DecryptAuthRefreshToken(ctxName string) error {
	if (strings.HasPrefix(c.AuthRefreshToken, secret.AesGcm) && c.Salt != nil) || (c.AuthRefreshToken != "" && runtime.GOOS == "windows") {
		decryptedAuthRefreshToken, err := secret.Decrypt(ctxName, c.AuthRefreshToken, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.AuthRefreshToken = decryptedAuthRefreshToken
	}

	return nil
}

func (c *ContextState) IsExpired() bool {
	if c == nil {
		return false
	}

	token, err := jwt.ParseSigned(c.AuthToken)
	if err != nil {
		return false
	}

	var claims map[string]any
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return false
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return false
	}

	return float64(time.Now().Unix()) > exp
}
