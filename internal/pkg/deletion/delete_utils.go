package deletion

import (
	"fmt"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func ValidateArgsForDeletion(cmd *cobra.Command, args []string, resourceType string, callDescribeEndpoint func(string) error) error {
	var invalidArgs []string
	for _, arg := range args {
		if err := callDescribeEndpoint(arg); err != nil {
			invalidArgs = append(invalidArgs, arg)
		}
	}

	if len(invalidArgs) != 0 {
		NotFoundErrorMsg := `%s "%s" not found`
		invalidArgsErrMsg := fmt.Sprintf(NotFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		if len(invalidArgs) > 1 {
			invalidArgsErrMsg = fmt.Sprintf(NotFoundErrorMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		}
		return errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.ListResourceSuggestions, resourceType, pcmd.FullParentName(cmd)))
	}

	return nil
}

func PrintSuccessfulDeletionMsg(successful []string, resourceType string) {
	if len(successful) == 1 {
		output.Printf(errors.DeletedResourceMsg, resourceType, successful[0])
	} else if len(successful) > 1 {
		output.Printf(errors.DeletedResourcesMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(successful, "and"))
	}
}
