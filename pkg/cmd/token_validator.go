package cmd

import (
	"fmt"
	"time"

	"github.com/go-jose/go-jose/v3/jwt"
	"github.com/jonboulle/clockwork"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/version"
)

type JWTValidator interface {
	Validate(context *config.Context) error
}

type JWTValidatorImpl struct {
	Clock   clockwork.Clock
	Version *version.Version
}

func NewJWTValidator() *JWTValidatorImpl {
	return &JWTValidatorImpl{
		Clock: clockwork.NewRealClock(),
	}
}

// Validate returns an error if the JWT in the specified context is invalid.
// The JWT is invalid if it's not parsable or expired.
func (v *JWTValidatorImpl) Validate(context *config.Context) error {
	token, err := jwt.ParseSigned(context.GetAuthToken())
	if err != nil {
		return new(ccloudv1.InvalidTokenError)
	}

	var claims map[string]any
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}

	exp, ok := claims["exp"].(float64)
	if !ok {
		return fmt.Errorf("malformed token: no expiration")
	}

	// Add a time buffer of 1 minute to the token validator
	if float64(v.Clock.Now().Add(time.Minute).Unix()) > exp {
		return errors.NewErrorWithSuggestions(errors.ExpiredTokenErrorMsg, errors.ExpiredTokenSuggestions)
	}

	return nil
}
