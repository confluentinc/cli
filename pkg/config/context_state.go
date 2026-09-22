package config

import (
	"regexp"
	"runtime"
	"time"

	"github.com/go-jose/go-jose/v4/jwt"

	"github.com/confluentinc/cli/v4/pkg/jose"
	"github.com/confluentinc/cli/v4/pkg/secret"
)

const (
	authTokenRegex        = `^[\w-]*\.[\w-]*\.[\w-]*$`
	authRefreshTokenRegex = `^(v1\..*)$`
)

type ContextState struct {
	// Deprecated
	Auth *AuthConfig `json:"auth,omitempty"`

	AuthToken        string `json:"-"`
	AuthRefreshToken string `json:"-"`
	Salt             []byte `json:"-"`
	Nonce            []byte `json:"-"`
}

func (c *ContextState) DecryptAuthToken(ctxName string) error {
	reg := regexp.MustCompile(authTokenRegex)
	if !reg.MatchString(c.AuthToken) && c.AuthToken != "" && (c.Salt != nil || runtime.GOOS == "windows") {
		decryptedAuthToken, err := secret.Decrypt(ctxName, c.AuthToken, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.AuthToken = decryptedAuthToken
	}

	return nil
}

func (c *ContextState) DecryptAuthRefreshToken(ctxName string) error {
	reg := regexp.MustCompile(authRefreshTokenRegex)
	if !reg.MatchString(c.AuthRefreshToken) && c.AuthRefreshToken != "" && (c.Salt != nil || runtime.GOOS == "windows") {
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

	token, err := jwt.ParseSigned(c.AuthToken, jose.SignatureAlgorithms)
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
