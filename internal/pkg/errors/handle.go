package errors

import (
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
	err = catchCCloudTokenErrors(err)
	err = catchCCloudBackendUnmarshallingError(err)
	err = catchTypedErrors(err)
	err = catchMDSErrors(err)
	err = catchCoreV1Errors(err)
	err = catchOpenAPIError(err)
	return err
}
