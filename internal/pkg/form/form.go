package form

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/output"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

/*
A multi-question form. Examples:

* Signup
First Name: Brian
Last Name: Strauch
Email: bstrauch@confluent.io

* Login
Username: user
Password: ****

* Confirmation
Submit? (y/n): y

* Default Values
Save file as: (file.txt) other.txt
*/

type Form struct {
	Fields    []Field
	Responses map[string]any
}

func New(fields ...Field) *Form {
	return &Form{
		Fields:    fields,
		Responses: make(map[string]any),
	}
}

func (f *Form) Prompt(prompt Prompt) error {
	for i := 0; i < len(f.Fields); i++ {
		field := f.Fields[i]
		output.Print(field.String())

		val, err := field.read(prompt)
		if err != nil {
			return err
		}

		res, err := field.validate(val)
		if err != nil {
			if fmt.Sprintf(errors.InvalidInputFormatErrorMsg, val, field.ID) == err.Error() {
				output.ErrPrintln(err)
				i-- // re-prompt on invalid regex
				continue
			}
			return err
		}
		if checkRequiredYes(field, res) {
			i-- // re-prompt on required yes
		}

		f.Responses[field.ID] = res
	}

	return nil
}

func ConfirmDeletionWithString(cmd *cobra.Command, promptMsg, stringToType string) error {
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return err
	} else if force {
		return nil
	}

	prompt := NewPrompt()
	f := New(Field{ID: "confirm", Prompt: promptMsg})
	if err := f.Prompt(prompt); err != nil {
		return err
	}

	if f.Responses["confirm"].(string) == stringToType || f.Responses["confirm"].(string) == fmt.Sprintf(`"%s"`, stringToType) {
		return nil
	}

	DeleteResourceConfirmSuggestions := "Use the `--force` flag to delete without a confirmation prompt."
	return errors.NewErrorWithSuggestions(fmt.Sprintf(`input does not match "%s"`, stringToType), DeleteResourceConfirmSuggestions)
}

func DefaultPromptString(resourceType, id, stringToType string) string {
	return fmt.Sprintf(errors.DeleteResourceConfirmMsg, resourceType, id, stringToType)
}

func ConfirmDeletionYesNo(cmd *cobra.Command, promptMsg string) (bool, error) {
	if force, err := cmd.Flags().GetBool("force"); err != nil {
		return false, err
	} else if force {
		return true, nil
	}

	prompt := NewPrompt()
	f := New(Field{ID: "confirm", Prompt: promptMsg, IsYesOrNo: true})
	if err := f.Prompt(prompt); err != nil {
		return false, errors.New(errors.FailedToReadInputErrorMsg)
	}

	return f.Responses["confirm"].(bool), nil
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

func ConfirmEnter() error {
	// This function prevents echoing of user input instead of displaying text or *'s by using
	// term.ReadPassword so that the CLI will appear to wait until 'enter' or 'Ctrl-C' are entered.
	output.Print("Press enter to continue or Ctrl-C to cancel:")

	if _, err := term.ReadPassword(int(os.Stdin.Fd())); err != nil {
		return err
	}
	// Warning: do not remove this print line; it prevents an unexpected interaction with browser.OpenUrl causing pages to open in the background
	output.Print("\n")

	return nil
}

func checkRequiredYes(field Field, res any) bool {
	if field.IsYesOrNo && field.RequireYes && !res.(bool) {
		output.Println("You must accept to continue. To abandon flow, use Ctrl-C.")
		return true
	}
	return false
}
