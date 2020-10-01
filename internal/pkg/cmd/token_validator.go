package cmd

import (
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/jonboulle/clockwork"
	"gopkg.in/square/go-jose.v2/jwt"

	v3 "github.com/confluentinc/cli/internal/pkg/config/v3"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"github.com/confluentinc/cli/internal/pkg/version"
)

type JWTValidator interface {
	Validate(context *v3.Context) error
}

type JWTValidatorImpl struct {
	Logger  *log.Logger
	Clock   clockwork.Clock
	Version *version.Version
}

func NewJWTValidator(logger *log.Logger) *JWTValidatorImpl {
	return &JWTValidatorImpl{
		Logger: logger,
		Clock:  clockwork.NewRealClock(),
	}
}

// Validate returns an error if the JWT in the specified context is invalid.
// The JWT is invalid if it's not parsable or expired.
func (v *JWTValidatorImpl) Validate(context *v3.Context) error {
	var authToken string
	if context != nil {
		authToken = context.State.AuthToken
	}
	var claims map[string]interface{}
	token, err := jwt.ParseSigned(authToken)
	if err != nil {
		return new(ccloud.InvalidTokenError)
	}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		return err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		return errors.New(errors.MalformedJWTNoExprErrorMsg)
	}
	if float64(v.Clock.Now().Unix()) > exp {
		v.Logger.Debug("Token expired.")
		return new(ccloud.ExpiredTokenError)
	}
	return nil
}
