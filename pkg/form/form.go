package form

import (
	"fmt"
	"os"

	"golang.org/x/term"

	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/output"
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
				output.ErrPrintf("Error: %v\n", err)
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

func ConfirmEnter() error {
	// This function prevents echoing of user input instead of displaying text or *'s by using
	// term.ReadPassword so that the CLI will appear to wait until 'enter' or 'Ctrl-C' are entered.
	output.Print("Press enter to continue or Ctrl-C to cancel:")

	if _, err := term.ReadPassword(int(os.Stdin.Fd())); err != nil {
		return err
	}
	// Warning: do not remove this print line; it prevents an unexpected interaction with browser.OpenUrl causing pages to open in the background
	output.Println("")

	return nil
}

func checkRequiredYes(field Field, res any) bool {
	if field.IsYesOrNo && field.RequireYes && !res.(bool) {
		output.Println("You must accept to continue. To abandon flow, use Ctrl-C.")
		return true
	}
	return false
}
