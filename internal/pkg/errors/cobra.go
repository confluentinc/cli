package errors

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/go-editor"
)

var messages = map[error]string{
	ErrNoContext:      "You must login to access Confluent Cloud.",
	ErrUnauthorized:   "You must login to access Confluent Cloud.",
	ErrExpiredToken:   "Your access to Confluent Cloud has expired. Please login again.",
	ErrIncorrectAuth:  "You have entered an incorrect username or password. Please try again.",
	ErrMalformedToken: "Your auth token has been corrupted. Please login again.",
	ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	ErrNotFound:       "Kafka cluster not found.", // TODO: parametrize ErrNotFound for better error messaging
	ErrNoKafkaContext: "You must pass --cluster or set an active kafka in your context with 'kafka cluster use'",
}

// HandleCommon provides standard error messaging for common errors.
func HandleCommon(err error, cmd *cobra.Command) error {
	// Give an indication of successful completion
	if err == nil {
		return nil
	}

	// Intercept errors to prevent usage from being printed.
	if msg, ok := messages[err]; ok {
		cmd.SilenceUsage = true
		return fmt.Errorf(msg)
	}

	switch err.(type) {
	case NotAuthenticatedError:
		cmd.SilenceUsage = true
		return err
	case UnknownKafkaContextError:
		cmd.SilenceUsage = true
		return fmt.Errorf("no auth found for Kafka %s, please run `ccloud kafka cluster auth` first", err.Error())
	case *UnconfiguredAPIKeyContextError:
		cmd.SilenceUsage = true
		return err
	// TODO: ErrEditing is declared incorrectly as "type ErrEditing error". That doesn't work for type switches, so put last
	case editor.ErrEditing:
		cmd.SilenceUsage = true
		return err
	}

	return err
}
