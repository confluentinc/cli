package errors

import (
	"fmt"
	"reflect"

	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"

	"github.com/confluentinc/go-editor"
	mds "github.com/confluentinc/mds-sdk-go"
)

var messages = map[error]string{
	ErrNoContext:      "You must login to run that command.",
	ErrNotLoggedIn:    "You must login to run that command.",
	ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
	ErrNoKafkaContext: "You must pass --cluster or set an active kafka in your context with 'kafka cluster use'",
}

var typeMessages = map[reflect.Type]string{
	reflect.TypeOf(&ccloud.InvalidLoginError{}): "You have entered an incorrect username or password. Please try again.",
	reflect.TypeOf(&ccloud.ExpiredTokenError{}): "Your session has expired. Please login again.",
	reflect.TypeOf(&ccloud.InvalidTokenError{}): "Your auth token has been corrupted. Please login again.",
}

// HandleCommon provides standard error messaging for common errors.
func HandleCommon(err error, cmd *cobra.Command) error {
	// Give an indication of successful completion
	if err == nil {
		return nil
	}

	if oerr, ok := err.(mds.GenericOpenAPIError); ok {
		cmd.SilenceUsage = true
		return fmt.Errorf(oerr.Error() + ": " + string(oerr.Body()))
	}

	// Intercept errors to prevent usage from being printed.
	if msg, ok := messages[err]; ok {
		cmd.SilenceUsage = true
		return fmt.Errorf(msg)
	}
	if msg, ok := typeMessages[reflect.TypeOf(err)]; ok {
		cmd.SilenceUsage = true
		return fmt.Errorf(msg)
	}

	switch e := err.(type) {
	case *UnspecifiedKafkaClusterError:
		cmd.SilenceUsage = true
		return fmt.Errorf("no auth found for Kafka %s, please run `ccloud kafka cluster auth` first", err.Error())
	case *UnspecifiedAPIKeyError:
		cmd.SilenceUsage = true
		return fmt.Errorf("no API key selected for %s, please select an api-key first (e.g., with `api-key use`)", e.ClusterID)
	case *UnconfiguredAPISecretError:
		cmd.SilenceUsage = true
		return err
	// TODO: ErrEditing is declared incorrectly as "type ErrEditing error"
	//  That doesn't work for type switches, so put last otherwise everything will hit this case
	case editor.ErrEditing:
		cmd.SilenceUsage = true
		return err
	}

	return err
}
