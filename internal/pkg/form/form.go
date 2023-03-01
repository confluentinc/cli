package form

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"github.com/confluentinc/cli/internal/pkg/errors"
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

func (f *Form) Prompt(command *cobra.Command, prompt Prompt) error {
	for i := 0; i < len(f.Fields); i++ {
		field := f.Fields[i]
		utils.Print(command, field.String())

		val, err := field.read(prompt)
		if err != nil {
			return err
		}

		res, err := field.validate(val)
		if err != nil {
			if fmt.Sprintf(errors.InvalidInputFormatErrorMsg, val, field.ID) == err.Error() {
				utils.ErrPrintln(command, err)
				i-- //re-prompt on invalid regex
				continue
			}
			return err
		}
		if checkRequiredYes(command, field, res) {
			i-- //re-prompt on required yes
		}

		f.Responses[field.ID] = res
	}

	return nil
}

func ConfirmDeletion(cmd *cobra.Command, promptMsg, stringToType string) (bool, error) {
	force, err := cmd.Flags().GetBool("force")
	if err != nil {
		return false, err
	}
	if force {
		return true, nil
	}

	prompt := NewPrompt(os.Stdin)
	isYesNo := stringToType == ""
	f := New(Field{ID: "confirm", Prompt: promptMsg, IsYesOrNo: isYesNo})
	if err := f.Prompt(cmd, prompt); err != nil && isYesNo {
		return false, errors.New(errors.FailedToReadInputErrorMsg)
	} else if err != nil {
		return false, err
	}

	if isYesNo {
		return f.Responses["confirm"].(bool), nil
	}

	if f.Responses["confirm"].(string) == stringToType || f.Responses["confirm"].(string) == fmt.Sprintf(`"%s"`, stringToType) {
		return true, nil
	}

	DeleteResourceConfirmSuggestions := "Use the `--force` flag to delete without a confirmation prompt."
	return false, errors.NewErrorWithSuggestions(fmt.Sprintf(`input does not match "%s"`, stringToType), DeleteResourceConfirmSuggestions)
}

func ConfirmEnter(cmd *cobra.Command) error {
	// This function prevents echoing of user input instead of displaying text or *'s by using
	// term.ReadPassword so that the CLI will appear to wait until 'enter' or 'Ctrl-C' are entered.
	utils.Print(cmd, "Press enter to continue or Ctrl-C to cancel:")

	if _, err := term.ReadPassword(int(os.Stdin.Fd())); err != nil {
		return err
	}
	// Warning: do not remove this print line; it prevents an unexpected interaction with browser.OpenUrl causing pages to open in the background
	utils.Print(cmd, "\n")

	return nil
}

func checkRequiredYes(cmd *cobra.Command, field Field, res any) bool {
	if field.IsYesOrNo && field.RequireYes && !res.(bool) {
		utils.Println(cmd, "You must accept to continue. To abandon flow, use Ctrl-C.")
		return true
	}
	return false
}
