package common

import (
	"fmt"

	"github.com/codyaray/go-editor"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/shared"
)

// HandleError provides standard error messaging for common errors.
func HandleError(err error, cmd *cobra.Command) error {
	out := cmd.OutOrStderr()
	switch err {
	case shared.ErrNoContext:
		fallthrough
	case shared.ErrUnauthorized:
		fmt.Fprintln(out, "You must login to access Confluent Cloud.")
	case shared.ErrExpiredToken:
		fmt.Fprintln(out, "Your access to Confluent Cloud has expired. Please login again.")
	case shared.ErrIncorrectAuth:
		fmt.Fprintln(out, "You have entered an incorrect username or password. Please try again.")
	case shared.ErrMalformedToken:
		fmt.Fprintln(out, "Your auth token has been corrupted. Please login again.")
	case shared.ErrNotImplemented:
		fmt.Fprintln(out, "Sorry, this functionality is not yet available in the CLI.")
	case shared.ErrNotFound:
		fmt.Fprintln(out, "Kafka cluster not found.") // TODO: parametrize ErrNotFound for better error messaging
	default:
		switch err.(type) {
		case editor.ErrEditing:
			fmt.Fprintln(out, err)
		case shared.NotAuthenticatedError:
			fmt.Fprintln(out, err)
		case shared.KafkaError:
			fmt.Fprintln(out, err)
		default:
			return err
		}
	}
	return nil
}
