package errors

import (
	"fmt"
	"io"
	"reflect"
	"strings"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

var (
	suggestionsMessageFormat = "\nSuggestions:\n    %s\n"

    messages = map[error]string{
		ErrNoContext:      UserNotLoggedInErrMsg,
		ErrNotLoggedIn:    UserNotLoggedInErrMsg,
		ErrNotImplemented: "Sorry, this functionality is not yet available in the CLI.",
		ErrNoKafkaContext: "You must pass --cluster or set an active kafka in your context with 'kafka cluster use'",
	}
)

func HandleCommon(err error, cmd *cobra.Command) error {
	cmd.SilenceUsage = true
	e := HandleCCloudSDKGoError(err)
	if e != nil {
		return e
	}
	e = HandleSaramaError(err)
	if e != nil {
		return e
	}

	// [CLI-505] mds.GenericOpenAPIErrors are not hashable so messages[err] panics;
	// so check if the error is hashable before trying to use messages[err]
	// (This is a recommended way of checking whether a variable is hashable, see
	//  https://groups.google.com/forum/#!topic/golang-nuts/fpzQdHBdV3c )
	k := reflect.TypeOf(err).Kind()
	hashable := k < reflect.Array || k == reflect.Ptr || k == reflect.UnsafePointer
	if hashable {
		if msg, ok := messages[err]; ok {
			return fmt.Errorf(msg)
		}
	}
	switch e := err.(type) {
	case mds.GenericOpenAPIError:
		return fmt.Errorf("metadata service backend error: " + e.Error() + ": " + string(e.Body()))
	case *corev1.Error:
		var result error
		result = multierror.Append(result, e)
		for name, msg := range e.GetNestedErrors() {
			result = multierror.Append(result, fmt.Errorf("%s: %s", name, msg))
		}
		return result
	case *UnspecifiedAPIKeyError:
		return fmt.Errorf("no API key selected for %s, please select an api-key first (e.g., with `api-key use`)", e.ClusterID)
	case *UnspecifiedCredentialError:
		// TODO: Add more context to credential error messages (add variable error).
		return fmt.Errorf(ConfigUnspecifiedCredentialError, e.ContextName)
	case *UnspecifiedPlatformError:
		// TODO: Add more context to platform error messages (add variable error).
		return fmt.Errorf(ConfigUnspecifiedPlatformError, e.ContextName)
	}


	return err
}

func HandleSuggestionsMessageDisplay(err error, writer io.Writer) {
	cliErr, ok := err.(ErrorWithSuggestions)
	if ok && cliErr.GetSuggestionsMsg() != "" {
		_, _ = fmt.Fprintf(writer, suggestionsMessageFormat, cliErr.GetSuggestionsMsg())
	}
}

func HandleCCloudSDKGoError(err error) error {
	switch err.(type) {
	case *ccloud.InvalidLoginError:
		return fmt.Errorf("You have entered an incorrect username or password. Please try again.")
	case *ccloud.InvalidTokenError:
		return fmt.Errorf(CorruptedAuthTokenErrorMsg)
	case *ccloud.ExpiredTokenError:
		return fmt.Errorf("expired token")
	}

	errMsg := err.Error()
	if strings.Contains(errMsg, "resource not found") {
		return NewErrorWithSuggestions("not found resource", "check your resource name and stuff dude")
	} else if strings.Contains(errMsg,"logicalCluster: Authentication failed") {
		return NewErrorWithSuggestions("not ready", "wait a bit fam!")
	}

	// non existent topic produce and consume
	// Failed to produce offset -1: kafka server: Request was for a topic or partition that does not exist on this broker.

	return nil
}

func HandleSaramaError(err error) error {
	if strings.Contains(err.Error(), "client has run out of available brokers to talk to (Is your cluster reachable?)") {
		return NewErrorWithSuggestions("Unable to connect to kafka cluster", "Check your api key and secret")
	}
	return nil
}
