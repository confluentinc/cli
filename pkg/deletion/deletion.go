package deletion

import (
	"fmt"
	"os"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func ValidateAndConfirmDeletionYesNo(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	return PromptYesOrNo(cmd, DefaultYesNoDeletePromptString(resourceType, args))
}

func ValidateAndConfirmDeletion(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType, name string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	if len(args) > 1 {
		return PromptYesOrNo(cmd, DefaultYesNoDeletePromptString(resourceType, args))
	}

	promptString := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resourceType, args[0], name)
	if err := ConfirmDeletionWithString(cmd, promptString, name); err != nil {
		return err
	}

	return nil
}

func PromptYesOrNo(cmd *cobra.Command, promptMsg string) error {
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

func ConfirmDeletionWithString(cmd *cobra.Command, promptMsg, stringToType string) error {
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return err
	} else if force {
		return nil
	}

	prompt := form.NewPrompt()
	f := form.New(form.Field{ID: "confirm", Prompt: promptMsg})
	if err := f.Prompt(prompt); err != nil {
		return err
	}

	if f.Responses["confirm"].(string) == stringToType || f.Responses["confirm"].(string) == fmt.Sprintf(`"%s"`, stringToType) {
		return nil
	}

	return errors.NewErrorWithSuggestions(
		fmt.Sprintf(`input does not match "%s"`, stringToType),
		"Use the `--force` flag to delete without a confirmation prompt.",
	)
}

func DeleteWithoutMessage(args []string, callDeleteEndpoint func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deletedIds []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, fmt.Errorf("failed to delete %s: %w", id, err))
		} else {
			deletedIds = append(deletedIds, id)
		}
	}

	return deletedIds, errs.ErrorOrNil()
}

func Delete(args []string, callDeleteEndpoint func(string) error, resourceType string) ([]string, error) {
	deletedIds, err := DeleteWithoutMessage(args, callDeleteEndpoint)

	DeletedResourceMsg := "Deleted %s %s.\n"
	if len(deletedIds) == 1 {
		output.Printf(false, DeletedResourceMsg, resourceType, fmt.Sprintf(`"%s"`, deletedIds[0]))
	} else if len(deletedIds) > 1 {
		output.Printf(false, DeletedResourceMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(deletedIds, "and"))
	}

	return deletedIds, err
}

func DefaultYesNoDeletePromptString(resourceType string, idList []string) string {
	var promptMsg string
	if len(idList) == 1 {
		promptMsg = fmt.Sprintf(`Are you sure you want to delete %s "%s"?`, resourceType, idList[0])
	} else {
		promptMsg = fmt.Sprintf("Are you sure you want to delete %ss %s?", resourceType, utils.ArrayToCommaDelimitedString(idList, "and"))
	}

	return promptMsg
}

func ValidateAndConfirmUndeletion(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType, name string) error {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return err
	}

	if len(args) > 1 {
		return PromptYesOrNo(cmd, DefaultYesNoUndeletePromptString(resourceType, args))
	}

	promptString := fmt.Sprintf(errors.UndeleteResourceConfirmMsg, resourceType, args[0], name)
	if err := ConfirmDeletionWithString(cmd, promptString, name); err != nil {
		return err
	}

	return nil
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
		output.Printf(false, UndeletedResourceMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(undeletedIds, "and"))
	}

	return undeletedIds, err
}

func DefaultYesNoUndeletePromptString(resourceType string, idList []string) string {
	var promptMsg string
	if len(idList) == 1 {
		promptMsg = fmt.Sprintf(`Are you sure you want to undelete %s "%s"?`, resourceType, idList[0])
	} else {
		promptMsg = fmt.Sprintf("Are you sure you want to undelete %ss %s?", resourceType, utils.ArrayToCommaDelimitedString(idList, "and"))
	}

	return promptMsg
}
