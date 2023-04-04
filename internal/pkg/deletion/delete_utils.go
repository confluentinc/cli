package deletion

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func ValidateArgsForDeletion(cmd *cobra.Command, args []string, resourceType string, callDescribeEndpoint func(string) error) ([]string, error) {
	var validArgs, invalidArgs []string
	for _, arg := range args {
		if err := callDescribeEndpoint(arg); err != nil {
			invalidArgs = append(invalidArgs, arg)
		} else {
			validArgs = append(validArgs, arg)
		}
	}

	var invalidArgsErrMsg string
	if len(invalidArgs) == 1 {
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
	} else if len(invalidArgs) > 1 {
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
	}

	if len(invalidArgs) != 0 {
		if warn, err := cmd.Flags().GetBool("skip-invalid"); err != nil {
			return nil, err
		} else if warn {
			output.ErrPrintln(invalidArgsErrMsg)
			return validArgs, nil
		}

		if len(validArgs) == 1 {
			return nil, errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.DeleteNotFoundSuggestions, resourceType))
		} else if len(validArgs) > 1 {
			return nil, errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.DeleteNotFoundSuggestions, resource.Plural(resourceType)))
		} else {
			return nil, errors.New(invalidArgsErrMsg)
		}
	}

	return args, nil
}

func PrintSuccessfulDeletionMsg(successful []string, resourceType string) {
	if len(successful) == 1 {
		output.Printf(errors.DeletedResourceMsg, resourceType, successful[0])
	} else if len(successful) > 1 {
		output.Printf(errors.DeletedResourcesMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(successful, "and"))
	}
}
