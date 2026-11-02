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

	// tokenExpirySafetyMargin is how long before the "exp" claim a token is treated as expired.
	tokenExpirySafetyMargin = 60 * time.Second
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

// IsExpired reports whether the session token should be treated as expired.
// A token is considered expired once the current time is within
// tokenExpirySafetyMargin of its "exp" claim, so that a token is not used
// moments before the server would reject it.
//
// A nil state, an unparseable token, or a token without a numeric "exp" claim
// is NOT treated as expired: we cannot prove expiry, and the only caller
// (the 401 handler in ccloudv2) uses this to decide whether to attempt a
// session refresh, so reporting "expired" for opaque tokens would trigger
// refreshes that cannot be validated.
func (c *ContextState) IsExpired() bool {
	return c.isExpiredAt(time.Now())
}

func (c *ContextState) isExpiredAt(now time.Time) bool {
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

	return float64(now.Unix()) >= exp-tokenExpirySafetyMargin.Seconds()
}
