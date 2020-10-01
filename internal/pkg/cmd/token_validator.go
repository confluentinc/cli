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

func NewJWTValidator(logger *log.Logger, clock clockwork.Clock) *JWTValidatorImpl {
	return &JWTValidatorImpl{
		Logger: logger,
		Clock:  clock,
	}
}

// validate token (not expired)
func (v *JWTValidatorImpl) Validate(context *v3.Context) error {
	// validate token (not expired)
	var authToken string
	if context != nil {
		authToken = context.State.AuthToken
	}
	var claims map[string]interface{}
	token, err := jwt.ParseSigned(authToken)
	if err != nil {
		//return v.updateToken(new(ccloud.InvalidTokenError), context)
		return err
	}
	if err := token.UnsafeClaimsWithoutVerification(&claims); err != nil {
		//return v.updateToken(err, context)
		return err
	}
	exp, ok := claims["exp"].(float64)
	if !ok {
		//return v.updateToken(errors.New(errors.MalformedJWTNoExprErrorMsg), context)
		return errors.New(errors.MalformedJWTNoExprErrorMsg)
	}
	if float64(v.Clock.Now().Unix()) > exp {
		v.Logger.Debug("Token expired.")
		//return v.updateToken(new(ccloud.ExpiredTokenError), context)
		return new(ccloud.ExpiredTokenError)
	}
	return nil
}

//func (v *JWTValidatorImpl) updateToken(tokenError error, context *v3.Context) error {
//	if context == nil {
//		v.Logger.Debug("Context is nil. Cannot attempt to update auth token.")
//		return tokenError
//	}
//	var updateErr error
//	if v.CLIName == "ccloud" {
//		updateErr = v.UpdateTokenHandler.UpdateCCloudAuthTokenUsingNetrcCredentials(context, v.Version.UserAgent, v.Logger)
//	} else {
//		updateErr = v.UpdateTokenHandler.UpdateConfluentAuthTokenUsingNetrcCredentials(context, v.Logger)
//	}
//	if updateErr == nil {
//		return nil
//	}
//	return tokenError
//}
