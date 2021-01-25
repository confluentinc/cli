package form

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

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
	Responses map[string]interface{}
}

type Field struct {
	ID           string
	Prompt       string
	DefaultValue interface{}
	IsYesOrNo    bool
	IsHidden     bool
	Regex        string
	RequireYes   bool
}

func New(fields ...Field) *Form {
	return &Form{
		Fields:    fields,
		Responses: make(map[string]interface{}),
	}
}

func (f *Form) Prompt(command *cobra.Command, prompt Prompt) error {
	for i := 0; i < len(f.Fields); i++ {
		field := f.Fields[i]
		show(command, field)

		val, err := read(field, prompt)
		if err != nil {
			return err
		}

		res, err := validate(field, val)
		if err != nil {
			if fmt.Sprintf(errors.InvalidInputFormatMsg, val, field.ID) == err.Error() {
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

func show(cmd *cobra.Command, field Field) {
	utils.Print(cmd, field.Prompt)
	if field.IsYesOrNo {
		utils.Print(cmd, " (y/n)")
	}
	utils.Print(cmd, ": ")

	if field.DefaultValue != nil {
		utils.Printf(cmd, "(%v) ", field.DefaultValue)
	}
}

func read(field Field, prompt Prompt) (string, error) {
	var val string
	var err error

	if field.IsHidden {
		val, err = prompt.ReadLineMasked()
	} else {
		val, err = prompt.ReadLine()
	}
	if err != nil {
		return "", err
	}

	return val, nil
}

func validate(field Field, val string) (interface{}, error) {
	if field.IsYesOrNo {
		switch strings.ToUpper(val) {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}
		return false, fmt.Errorf(errors.InvalidChoiceMsg, val)
	}

	if val == "" && field.DefaultValue != nil {
		return field.DefaultValue, nil
	}

	if field.Regex != "" {
		field_rgx, _ := regexp.Compile(field.Regex)
		if regex_match := field_rgx.MatchString(val); !regex_match {
			return nil, fmt.Errorf(errors.InvalidInputFormatMsg, val, field.ID)
		}
	}

	return val, nil
}

func checkRequiredYes(cmd *cobra.Command, field Field, res interface{}) bool {
	if field.IsYesOrNo && field.RequireYes && !res.(bool) {
		utils.Println(cmd, "You must accept to continue. To abandon flow, use Ctrl-C.")
		return true
	}
	return false
}
