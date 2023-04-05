package deletion

import (
	"fmt"

	"github.com/spf13/cobra"

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
		invalidArgsErrMsg := fmt.Sprintf(errors.NotFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		if len(invalidArgs) > 1 {
			invalidArgsErrMsg = fmt.Sprintf(errors.NotFoundErrorMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		}
		return errors.New(invalidArgsErrMsg)
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
