package http

import (
	"fmt"

	"google.golang.org/grpc/status"
)

/*
 * Invariants:
 * - Confluent SDK (http package) always returns an ApiError.
 * - Plugins always return an HTTP Error constant (top of this file)
 *
 * Error Flow:
 * - API error responses (json) are parsed into ApiError objects.
 *   - Note: API returns 404s for unauthorized resources, so HTTP package has to remap 404 -> 401 where appropriate.
 * - Plugins call ConvertAPIError() to transforms ApiError into HTTP Error constants
 * - GRPC encodes errors into Status objects when sent over the wire
 * - Commands call ConvertGRPCError() to transform these back into HTTP Error constants
 */

var (
	ErrIncorrectAuth  = fmt.Errorf("incorrect auth")
	ErrUnauthorized   = fmt.Errorf("unauthorized")
	ErrExpiredToken   = fmt.Errorf("expired")
	ErrMalformedToken = fmt.Errorf("malformed")
)

// TODO: reuse corev1.Error from cc-structs
type apiError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// API replies all have "error" field. Only set if non-successful HTTP response code.
type ApiError struct {
	Err     *apiError `json:"error"`
}

func (e *ApiError) Error() string {
	return fmt.Sprintf("confluent (%v): %v", e.Err.Code, e.Err.Message)
}

func (e *ApiError) OrNil() error {
	if e.Err != nil {
		return e
	}
	return nil
}

func ConvertAPIError(err error) error {
	if e, ok := err.(*ApiError); ok {
		switch e.Err.Message {
		// these messages are returned by the API itself
		case "token is expired":
			return ErrExpiredToken
		case "malformed token":
			return ErrMalformedToken
		// except this one.. its the special case of errUnauthorized from http/auth.go
		case "unauthorized":
			return ErrUnauthorized
		// TODO: assert invariant for default case: we're missing an ApiError -> HTTP Error constant mapping
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
