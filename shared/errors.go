package shared

import (
	"io"
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/pkg/errors"
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
 *  - Note: API returns 404s for unauthorized resources, so HTTP package has to remap 404 -> 401 where appropriate.
 * - Plugins call ConvertAPIError() to transforms corev1.Error into HTTP Error constants
 * - GRPC encodes errors into Status objects when sent over the wire
 * - Commands call ConvertGRPCError() to transform these back into HTTP Error constants
 */

type NotAuthenticatedError error
type KafkaError error

var (
	ErrNotImplemented = fmt.Errorf("not implemented")
	ErrIncorrectAuth  = fmt.Errorf("incorrect auth")
	ErrUnauthorized   = fmt.Errorf("unauthorized")
	ErrExpiredToken   = fmt.Errorf("expired")
	ErrMalformedToken = fmt.Errorf("malformed")
	ErrNotFound       = fmt.Errorf("not found")
	ErrNoContext      = fmt.Errorf("context not set")
	ErrBadRequest     = fmt.Errorf("malformed request")
)

type KafkaAPIError struct {
	Field string `json:"field"`
	Msg   string `json:"message"`
}

func (e *KafkaAPIError) Error() string {
	return e.Msg
}

// ConvertAPIError transforms a corev1.Error into one of the standard errors if it matches.
func ConvertAPIError(err error) error {
	if e, ok := errors.Cause(err).(*corev1.Error); ok {
		switch e.Message {
		// these messages are returned by the API itself
		case "token is expired":
			return ErrExpiredToken
		case "malformed token":
			return ErrMalformedToken
		// except this one.. its the special case of errUnauthorized from http/auth.go
		case "unauthorized":
			return ErrUnauthorized
		// except this one.. its the special case of errNotFound from http/client.go
		case "cluster not found":
			return ErrNotFound
			// TODO: assert invariant for default case: we're missing an corev1.Error -> HTTP Error constant mapping
		}
	}
	return err
}

// ConvertGRPCError unboxes and returns the underlying standard error sent over gRPC if it matches.
func ConvertGRPCError(err error) error {
	if err == nil {
		return nil
	}

	if s, ok := status.FromError(err); ok {
		switch s.Message() {
		case ErrExpiredToken.Error():
			return ErrExpiredToken
		case ErrMalformedToken.Error():
			return ErrMalformedToken
		case ErrUnauthorized.Error():
			return ErrUnauthorized
		case ErrNotFound.Error():
			return ErrNotFound
			// TODO: assert invariant for default case: we're missing a GRPC -> HTTP Error constant mapping
		}
		return fmt.Errorf(s.Message())
	}
	return err
}

// readBody returns the contents of reader r.
// The caller is still responsible for closing reader.
func readBody(r io.Reader) []byte {
	payload, _ := ioutil.ReadAll(r)
	return payload
}

// HandleKafkaAPIError returns an instance of KafkaAPIError
// Java Reference:
// https://github.com/confluentinc/blueway/blob/master/control-center/src/main/java/io/confluent/controlcenter/rest/KafkaExceptionMapper.java
func HandleKafkaAPIError(resp *http.Response, err error) error {
	if err != nil {
		return corev1.WrapErr(err, "an error occurred handling a request ")
	}

	switch {
	case resp.StatusCode >= http.StatusOK && resp.StatusCode < http.StatusMultipleChoices:
		break
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrExpiredToken
	default:
		var err KafkaAPIError
		body := readBody(resp.Body)
		if json.Valid(body) {
			if err := json.Unmarshal(body, &err); err != nil {
				return corev1.WrapErr(err, "")
			}
			return &err
		}
		return &KafkaAPIError{
			Field: "Unexpected HTTP error",
			Msg:   string(body),
		}
	}
	return nil
}
