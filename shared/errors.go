package shared

import (
	"fmt"

	"google.golang.org/grpc/status"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
)

/*
 * Invariants:
 * - Confluent SDK (http package) always returns a corev1.Error.
 * - Plugins always return an HTTP Error constant (top of this file)
 *
 * Error Flow:
 * - API error responses (json) are parsed into corev1.Error objects.
 *   - Note: API returns 404s for unauthorized resources, so HTTP package has to remap 404 -> 401 where appropriate.
 * - Plugins call ConvertAPIError() to transforms corev1.Error into HTTP Error constants
 * - GRPC encodes errors into Status objects when sent over the wire
 * - Commands call ConvertGRPCError() to transform these back into HTTP Error constants
 */

var (
	ErrNotImplemented = fmt.Errorf("not implemented")
	ErrIncorrectAuth  = fmt.Errorf("incorrect auth")
	ErrUnauthorized   = fmt.Errorf("unauthorized")
	ErrExpiredToken   = fmt.Errorf("expired")
	ErrMalformedToken = fmt.Errorf("malformed")
)

func ConvertAPIError(err error) error {
	if e, ok := err.(*corev1.Error); ok {
		switch e.Message {
		// these messages are returned by the API itself
		case "token is expired":
			return ErrExpiredToken
		case "malformed token":
			return ErrMalformedToken
		// except this one.. its the special case of errUnauthorized from http/auth.go
		case "unauthorized":
			return ErrUnauthorized
		// TODO: assert invariant for default case: we're missing an corev1.Error -> HTTP Error constant mapping
		}
	}
	return err
}

func ConvertGRPCError(err error) error {
	if s, ok := status.FromError(err); ok {
		// these messages are from the error constants at the top of this file
		switch s.Message() {
		case "expired":
			return ErrExpiredToken
		case "malformed":
			return ErrMalformedToken
		case "unauthorized":
			return ErrUnauthorized
		// TODO: assert invariant for default case: we're missing a GRPC -> HTTP Error constant mapping
		}
		return fmt.Errorf(s.Message())
	}
	return err
}
