package errors

import (
	"fmt"
	"github.com/confluentinc/ccloud-sdk-go"
	"github.com/spf13/cobra"
	"io"
)

var (
	directionsMessageFormat = "\nDirections:\n    %s\n"
)

func HandleCommon(err error, cmd *cobra.Command) error {
	if err == nil {
		return nil
	}
	var cliError CLIDefinedError
	switch e := err.(type) {
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

func HandleCCloudSDKGoError(err error) error {
	switch err.(type) {
	case *ccloud.InvalidLoginError:
		return fmt.Errorf("You have entered an incorrect username or password. Please try again.")
	case *ccloud.InvalidTokenError:
		return fmt.Errorf(CorruptedAuthTokenErrorMsg)
	case *ccloud.ExpiredTokenError:
		return fmt.Errorf("expired token")
	}

	switch err.Error() {
	case "resource not found":
		return err
	}
	switch "logicalCluster: Authentication failed" {
		return err
	}

	// non existent topic produce and consume


	// error for when no api-key for a resource

	//Error: kafka: client has run out of available brokers to talk to (Is your cluster reachable?)
	// This is what happens when your api key and secret is wrong

}
