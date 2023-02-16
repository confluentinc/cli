package v1

import (
	"regexp"

	"github.com/confluentinc/cli/internal/pkg/secret"
)

const (
	authTokenRegex        = `^[\w-]*\.[\w-]*\.[\w-]*$`
	authRefreshTokenRegex = `^(v1.*)$`
)

type ContextState struct {
	Auth             *AuthConfig `json:"auth" hcl:"auth"`
	AuthToken        string      `json:"auth_token" hcl:"auth_token"`
	AuthRefreshToken string      `json:"auth_refresh_token" hcl:"auth_refresh_token"`
	Salt             []byte      `json:"salt,omitempty"`
	Nonce            []byte      `json:"nonce,omitempty"`
}

func (c *ContextState) DecryptContextStateTokens(ctxName string) error {
	reg1 := regexp.MustCompile(authTokenRegex)
	if match := reg1.MatchString(c.AuthToken); !match && c.AuthToken != "" { // it's encrypted and not empty
		decryptedAuthToken, err := secret.Decrypt(ctxName, c.AuthToken, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.AuthToken = decryptedAuthToken
	}

	reg2 := regexp.MustCompile(authRefreshTokenRegex)
	if match := reg2.MatchString(c.AuthRefreshToken); !match && c.AuthRefreshToken != "" { // encrypted
		decryptedAuthRefreshToken, err := secret.Decrypt(ctxName, c.AuthRefreshToken, c.Salt, c.Nonce)
		if err != nil {
			return err
		}
		c.AuthRefreshToken = decryptedAuthRefreshToken
	}

	return nil
}
