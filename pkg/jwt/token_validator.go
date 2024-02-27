package jwt

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

type Validator interface {
	Validate(context *config.Context) error
}

type ValidatorImpl struct {
	Clock   clockwork.Clock
	Version *version.Version
}

func NewValidator() *ValidatorImpl {
	return &ValidatorImpl{
		Clock: clockwork.NewRealClock(),
	}
}

// Validate returns an error if the JWT in the specified context is invalid.
// The JWT is invalid if it's not parsable or expired.
func (v *ValidatorImpl) Validate(context *config.Context) error {
	expClaim, err := GetClaim(context.GetAuthToken(), "exp")
	if err != nil {
		return err
	}

	exp, ok := expClaim.(float64)
	if !ok {
		return fmt.Errorf("malformed token: exp claim has the wrong type")
	}

	// Add a time buffer of 1 minute to the token validator
	if float64(v.Clock.Now().Add(time.Minute).Unix()) > exp {
		return errors.NewErrorWithSuggestions(errors.ExpiredTokenErrorMsg, errors.ExpiredTokenSuggestions)
	}

	return nil
}

func GetClaim(jwtToken, claim string) (any, error) {
	token, err := jwt.ParseSigned(jwtToken)
	if err != nil {
		return nil, new(ccloudv1.InvalidTokenError)
	}

	var claims map[string]any
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return nil, err
	}

	val, ok := claims[claim]
	if !ok {
		return nil, fmt.Errorf("malformed token: no %s claim", claim)
	}

	return val, nil
}
