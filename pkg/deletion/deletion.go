package deletion

import (
	"fmt"

	"github.com/hashicorp/go-multierror"
	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
	"github.com/confluentinc/cli/v3/pkg/resource"
	"github.com/confluentinc/cli/v3/pkg/utils"
)

func ValidateAndConfirmDeletion(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType string) (bool, error) {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return false, err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return false, err
	}

	return ConfirmDeletionYesNo(cmd, DefaultYesNoPromptString(resourceType, args))
}

func ValidateAndConfirmDeletionWithName(cmd *cobra.Command, args []string, checkExistence func(string) bool, resourceType, name string) (bool, error) {
	if err := resource.ValidatePrefixes(resourceType, args); err != nil {
		return false, err
	}

	if err := resource.ValidateArgs(cmd, args, resourceType, checkExistence); err != nil {
		return false, err
	}

	if len(args) > 1 {
		return ConfirmDeletionYesNo(cmd, DefaultYesNoPromptString(resourceType, args))
	}

	promptString := fmt.Sprintf(errors.DeleteResourceConfirmMsg, resourceType, args[0], name)
	if err := ConfirmDeletionWithString(cmd, promptString, name); err != nil {
		return false, err
	}

	return true, nil
}

func ConfirmDeletionYesNo(cmd *cobra.Command, promptMsg string) (bool, error) {
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return false, err
	} else if force {
		return true, nil
	}

	prompt := form.NewPrompt()
	f := form.New(form.Field{ID: "confirm", Prompt: promptMsg, IsYesOrNo: true})
	if err := f.Prompt(prompt); err != nil {
		return false, errors.New(errors.FailedToReadInputErrorMsg)
	}

	return f.Responses["confirm"].(bool), nil
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

	DeleteResourceConfirmSuggestions := "Use the `--force` flag to delete without a confirmation prompt."
	return errors.NewErrorWithSuggestions(fmt.Sprintf(`input does not match "%s"`, stringToType), DeleteResourceConfirmSuggestions)
}

func delete(args []string, callDeleteEndpoint func(string) error) ([]string, error) {
	errs := &multierror.Error{ErrorFormat: errors.CustomMultierrorList}
	var deletedIDs []string
	for _, id := range args {
		if err := callDeleteEndpoint(id); err != nil {
			errs = multierror.Append(errs, err)
		} else {
			deletedIDs = append(deletedIDs, id)
		}
	}

	return deletedIDs, errs.ErrorOrNil()
}

func Delete(args []string, callDeleteEndpoint func(string) error, resourceType string) ([]string, error) {
	deletedIDs, err := delete(args, callDeleteEndpoint)

	DeletedResourceMsg := "Deleted %s %s.\n"
	if len(deletedIDs) == 1 {
		output.Printf(DeletedResourceMsg, resourceType, fmt.Sprintf("\"%s\"", deletedIDs[0]))
	} else if len(deletedIDs) > 1 {
		output.Printf(DeletedResourceMsg, resource.Plural(resourceType), utils.ArrayToCommaDelimitedString(deletedIDs, "and"))
	}

	return deletedIDs, err
}

func DeleteWithCustomMessage(args []string, callDeleteEndpoint func(string) error, singularMsg, pluralMsg string) ([]string, error) {
	deletedIDs, err := delete(args, callDeleteEndpoint)

	if len(deletedIDs) == 1 {
		output.Printf(singularMsg, fmt.Sprintf("\"%s\"", deletedIDs[0]))
	} else if len(deletedIDs) > 1 {
		output.Printf(pluralMsg, utils.ArrayToCommaDelimitedString(deletedIDs, "and"))
	}

	return deletedIDs, err
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
