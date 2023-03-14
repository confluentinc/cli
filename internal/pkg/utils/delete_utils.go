package utils

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/types"
)

func ValidateArgsForDeletion(cmd *cobra.Command, args []string, resourceType string, callListEndpoint func() (types.Set, error)) ([]string, error) {
	set, err := callListEndpoint()
	if err != nil {
		return nil, err
	}

	validArgs, invalidArgs := set.IntersectionAndDifference(args)
	var invalidArgsErrMsg string
	if len(invalidArgs) == 1 {
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resourceType, ArrayToCommaDelimitedStringWithAnd(invalidArgs))
	} else if len(invalidArgs) > 1 {
		invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resourceType), ArrayToCommaDelimitedStringWithAnd(invalidArgs))
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
