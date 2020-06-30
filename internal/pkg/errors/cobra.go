package errors

import (
	"fmt"
	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"io"
	"strings"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	"github.com/confluentinc/ccloud-sdk-go"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
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
	e := handleCCloudSDKGoError(err)
	if e != nil {
		return e
	}
	e = handleTypedErrors(err)
	if e != nil {
		return e
	}

	switch e := err.(type) {
	case mds.GenericOpenAPIError:
		return Errorf(GenericOpenAPIErrorMsg, e.Error(), string(e.Body()))
	case *corev1.Error:
		var result error
		result = multierror.Append(result, e)
		for name, msg := range e.GetNestedErrors() {
			result = multierror.Append(result, fmt.Errorf("%s: %s", name, msg))
		}
		return result
	}

	return err
}

func handleCCloudSDKGoError(err error) error {
	switch err.(type) {
	case *ccloud.InvalidLoginError:
		return NewErrorWithSuggestions(InvalidLoginErrorMsg, InvalidLoginSuggestions)
	case *ccloud.InvalidTokenError:
		return NewErrorWithSuggestions(CorruptedTokenErrorMsg, CorruptedTokenSuggestions)
	case *ccloud.ExpiredTokenError:
		return NewErrorWithSuggestions(ExpiredTokenErrorMsg, ExpiredTokenSuggestions)
	}
	return nil
}

func handleTypedErrors(err error) error {
	if typedErr, ok := err.(CLITypedError); ok {
		return typedErr.UserFacingError()
	}
	return err
}

func DisplaySuggestionsMessage(err error, writer io.Writer) {
	if err == nil {
		return
	}
	cliErr, ok := err.(ErrorWithSuggestions)
	if ok && cliErr.GetSuggestionsMsg() != "" {
		_, _ = fmt.Fprint(writer, composeSuggestionsMessage(cliErr.GetSuggestionsMsg()))
	}
}

func composeSuggestionsMessage(msg string) string {
	lines := strings.Split(msg, "\n")
	suggestionsMsg := suggestionsMessageHeader
	for _, line := range lines {
		suggestionsMsg += fmt.Sprintf(suggestionsLineFormat, line)
	}
	return suggestionsMsg
}

