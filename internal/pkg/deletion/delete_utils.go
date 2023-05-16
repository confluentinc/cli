package deletion

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/resource"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func ValidateArgs(cmd *cobra.Command, args []string, resourceType string, callDescribeEndpoint func(string) error) error {
	var invalidArgs []string
	for _, arg := range args {
		if err := callDescribeEndpoint(arg); err != nil {
			invalidArgs = append(invalidArgs, arg)
		}
	}

	if len(invalidArgs) != 0 {
		NotFoundErrorMsg := `%s %s not found`
		invalidArgsErrMsg := fmt.Sprintf(NotFoundErrorMsg, resourceType, utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		if len(invalidArgs) > 1 {
			invalidArgsErrMsg = fmt.Sprintf(NotFoundErrorMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(invalidArgs, "and"))
		}
		return errors.NewErrorWithSuggestions(invalidArgsErrMsg, fmt.Sprintf(errors.ListResourceSuggestions, resourceType, pcmd.FullParentName(cmd)))
	}

	return nil
}

func DefaultPostProcess(_ string) error {
	return nil
}

func DeleteResources(args []string, callDeleteEndpoint func(string) error, postProcess func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deleted []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deleted = append(deleted, id)
			if err := postProcess(id); err != nil {
				errs = multierror.Append(errs, err)
			}
		}
	}

	return deleted, errs.ErrorOrNil()
}

func PrintSuccessMsg(successful []string, resourceType string) {
	if len(successful) == 1 {
		output.Printf(errors.DeletedResourceMsg, resourceType, successful[0])
	} else if len(successful) > 1 {
		output.Printf("Deleted %s %s.\n", resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(successful, "and"))
	}
}

func DefaultPromptString(resourceType, id, stringToType string) string {
	return fmt.Sprintf(errors.DeleteResourceConfirmMsg, resourceType, id, stringToType)
}

func DefaultYesNoPromptString(resourceType string, idList []string) string {
	var promptMsg string
	if len(idList) == 1 {
		promptMsg = fmt.Sprintf(`Are you sure you want to delete %s "%s"?`, resourceType, idList[0])
	} else {
		promptMsg = fmt.Sprintf("Are you sure you want to delete %ss %s?", resourceType, utils.ArrayToCommaDelimitedString(idList, "and"))
	}

	return promptMsg
}
