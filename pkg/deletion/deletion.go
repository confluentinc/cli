package deletion

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/confluentinc/cli/v4/pkg/errors"
	"github.com/confluentinc/cli/v4/pkg/form"
	"github.com/confluentinc/cli/v4/pkg/output"
	"github.com/confluentinc/cli/v4/pkg/plural"
	"github.com/confluentinc/cli/v4/pkg/resource"
	"github.com/confluentinc/cli/v4/pkg/utils"
)

var pastTenseMap = map[string]string{
	"delete":    "deleted",
	"disable":   "disabled",
	"undelete":  "undeleted",
	"uninstall": "uninstalled",
}

func ValidateAndConfirm(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType string) error {
	return ValidateAndConfirmWithExtraWarning(cmd, args, checkExistence, resourceType, "")
}

func ValidateAndConfirmWithExtraWarning(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType string, extraWarning string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	return ConfirmPrompt(cmd, DefaultYesNoPromptString(cmd, resourceType, args, extraWarning))
}

func ConfirmPrompt(cmd *cobra.Command, promptMsg string) error {
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return err
	} else if force {
		return nil
	}

	prompt := form.NewPrompt()
	f := form.New(form.Field{ID: "confirm", Prompt: promptMsg, IsYesOrNo: true})
	if err := f.Prompt(prompt); err != nil {
		return fmt.Errorf(errors.FailedToReadInputErrorMsg)
	}

	if !f.Responses["confirm"].(bool) {
		os.Exit(0)
	}

	return nil
}

func DeleteWithoutMessage(cmd *cobra.Command, args []string, callDeleteEndpoint func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	operation := cmd.CalledAs()
	var deletedIds []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to %s %s: %w", operation, id, err))
		} else {
			deletedIds = append(deletedIds, id)
		}
	}

	return deletedIds, errs.ErrorOrNil()
}

func Delete(cmd *cobra.Command, args []string, callDeleteEndpoint func(string) error, resourceType string) ([]string, error) {
	deletedIds, err := DeleteWithoutMessage(cmd, args, callDeleteEndpoint)

	operation := cases.Title(language.Und).String(pastTenseMap[cmd.CalledAs()])
	DeletedResourceMsg := "%s %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(false, DeletedResourceMsg, operation, resourceType, fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(false, DeletedResourceMsg, operation, plural.Plural(resourceType), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return deletedIds, err
}

func DefaultYesNoPromptString(cmd *cobra.Command, resourceType string, idList []string, extraWarning string) string {
	operation := cmd.CalledAs()
	var promptMsg string
	if len(idList) == 1 {
		promptMsg = fmt.Sprintf(`Are you sure you want to %s %s "%s"?`, operation, resourceType, idList[0])
	} else {
		promptMsg = fmt.Sprintf("Are you sure you want to %s %s %s?", operation, plural.Plural(resourceType), utils.ArrayToCommaDelimitedString(idList, "and"))
	}
	promptMsg += extraWarning

	return promptMsg
}

func UndeleteWithoutMessage(args []string, callUndeleteEndpoint func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var undeletedIds []string
	for _, id := range args {
		if err := callUndeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to undelete %s: %w", id, err))
		} else {
			undeletedIds = append(undeletedIds, id)
		}
	}

	return undeletedIds, errs.ErrorOrNil()
}

func Undelete(args []string, callUndeleteEndpoint func(string) error, resourceType string) ([]string, error) {
	undeletedIds, err := UndeleteWithoutMessage(args, callUndeleteEndpoint)

	UndeletedResourceMsg := "Undeleted %s %s.\n"
	if len(undeletedIds) == 1 {
		output.Printf(false, UndeletedResourceMsg, resourceType, fmt.Sprintf(`"%s"`, undeletedIds[0]))
	} else if len(undeletedIds) > 1 {
		output.Printf(false, UndeletedResourceMsg, plural.Plural(resourceType), utils.ArrayToCommaDelimitedString(undeletedIds, "and"))
	}

	return undeletedIds, err
}
