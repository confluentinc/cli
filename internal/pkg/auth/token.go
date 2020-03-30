package auth

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/jonboulle/clockwork"
	"gopkg.in/square/go-jose.v2/jwt"
)

type TokenValidator interface {
	ValidateToken(token string) error
}

type JWTTokenValidator struct {
	Clock clockwork.Clock
}

// Validates a JWT token. Return nil if token is valid, and an error if it's not.
func (v *JWTTokenValidator) ValidateToken(token string) error {
	var claims map[string]interface{}
	jwtToken, err := jwt.ParseSigned(token)
	if err != nil {
		return err
	}
	if err := jwtToken.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}
	if exp, ok := claims["exp"].(float64); ok {
		if float64(v.Clock.Now().Unix()) > exp {
			return new(ccloud.ExpiredTokenError)
		}
	}
	return nil
}
