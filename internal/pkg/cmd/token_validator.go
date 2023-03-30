package cmd

import (
	"github.com/jonboulle/clockwork"
	"gopkg.in/square/go-jose.v2/jwt"

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type JWTValidator interface {
	Validate(context *v1.Context) error
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
func (v *JWTValidatorImpl) Validate(context *v1.Context) error {
	var authToken string
	if context != nil {
		authToken = context.State.AuthToken
	}
	var claims map[string]any
	token, err := jwt.ParseSigned(authToken)
	if err != nil {
		return new(ccloudv1.InvalidTokenError)
	}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New(errors.MalformedJWTNoExprErrorMsg)
	}
	if float64(v.Clock.Now().Unix()) > exp {
		log.CliLogger.Debug("Token expired.")
		return new(ccloudv1.ExpiredTokenError)
	}
	return nil
}
