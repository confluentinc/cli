package errors

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"io"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/confluentinc/mds-sdk-go"
)

var (
	messages = map[error]string{
		ErrNoContext:      UserNotLoggedInErrMsg,
		ErrNotLoggedIn:    UserNotLoggedInErrMsg,
		ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
		ErrNoKafkaContext: "You must pass --cluster or set an active kafka in your context with 'kafka cluster use'",
	}

	directionsMessageFormat = "\nDirections:\n    %s\n"
)

func HandleCommon(err error, cmd *cobra.Command) error {
	if err == nil {
		return nil
	}
	var cliError CLIDefinedError
	switch e := err.(type) {

	case mds.GenericOpenAPIError:
		cmd.SilenceUsage = true
		return fmt.Errorf(e.Error() + ": " + string(e.Body()))
	case *corev1.Error:
		var result error
		result = multierror.Append(result, e)
		for name, msg := range e.GetNestedErrors() {
			result = multierror.Append(result, fmt.Errorf("%s: %s", name, msg))
		}
		cmd.SilenceUsage = true
		return result
	case *ccloud.InvalidTokenError:
		cmd.SilenceUsage = true
		return fmt.Errorf(CorruptedAuthTokenErrorMsg)
	case CLIDefinedError:
		cliError = e
	default:
		cliError = NewUnexpectedCLIBehaviorErrorf(e.Error())
	}
	cmd.SilenceUsage = cliError.SilenceUsage()
	return cliError
}

func HandleDirectionsMessageDisplay(err error, writer io.Writer) {
	cliErr, ok := err.(CLIDefinedError)
	if ok && cliErr.GetDirectionsMsg() != "" {
		_, _ = fmt.Fprintf(writer, directionsMessageFormat, cliErr.GetDirectionsMsg())
	}
}
