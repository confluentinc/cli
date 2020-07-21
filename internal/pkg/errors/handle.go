package errors

import (
	"fmt"
	"io"
	"strings"

	"github.com/spf13/cobra"
)

var (
	suggestionsMessageHeader = "\nSuggestions:\n"
	suggestionsLineFormat    = "    %s\n"
)

func HandleCommon(err error, cmd *cobra.Command) error {
	if err == nil {
		return nil
	}
	cmd.SilenceUsage = true
	return handleErrors(err)
}

func handleErrors(err error) error {
	if err == nil {
		return nil
	}
	err = catchCCloudTokenErrors(err)
	err = catchCCloudBackendUnmarshallingError(err)
	err = catchTypedErrors(err)
	err = catchMDSErrors(err)
	err = catchCoreV1Errors(err)
	return err
}

func DisplaySuggestionsMessage(err error, writer io.Writer) {
	if err == nil {
		return
	}
	cliErr, ok := err.(ErrorWithSuggestions)
	if ok && cliErr.GetSuggestionsMsg() != "" {
		_, _ = fmt.Fprint(writer, ComposeSuggestionsMessage(cliErr.GetSuggestionsMsg()))
	}
}

func ComposeSuggestionsMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	suggestionsMsg := suggestionsMessageHeader
	for _, line := range lines {
		suggestionsMsg += fmt.Sprintf(suggestionsLineFormat, line)
	}
	return suggestionsMsg
}
