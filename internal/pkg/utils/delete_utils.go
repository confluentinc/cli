package utils

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
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
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resourceType, ArrayToCommaDelimitedString(invalidArgs, "and"))
	} else if len(invalidArgs) > 1 {
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resourceType), ArrayToCommaDelimitedString(invalidArgs, "and"))
	}

	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return nil, err
	} else if force && len(invalidArgs) > 0 {
		output.ErrPrintln(invalidArgsErrMsg)
		return validArgs, nil
	} else if len(invalidArgs) >= 1 {
		return nil, errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.DeleteNotFoundSuggestions, resourceType))
	}

	return args, nil
}
