package errors

import (
	"fmt"
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
