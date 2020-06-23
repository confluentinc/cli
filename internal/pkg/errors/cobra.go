package errors

import (
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"io"
	"strings"
)

var (
	directionsMessageFormat = "\nDirections:\n    %s\n"
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
	return err
}

func HandleSuggestionsMessageDisplay(err error, writer io.Writer) {
	cliErr, ok := err.(ErrorWithSuggestions)
	if ok && cliErr.GetSuggestionsMsg() != "" {
		_, _ = fmt.Fprintf(writer, directionsMessageFormat, cliErr.GetSuggestionsMsg())
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
