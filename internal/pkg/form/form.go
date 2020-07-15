package form

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/errors"
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
	Fields    map[string]Field
	Responses map[string]interface{}
}

type Field struct {
	Prompt       string
	DefaultValue interface{}
	IsYesOrNo    bool
	IsHidden     bool
}

func New(fields map[string]Field) *Form {
	return &Form{
		Fields:    fields,
		Responses: make(map[string]interface{}),
	}
}

func (f *Form) Prompt(command *cobra.Command, prompt cmd.Prompt) error {
	for id, field := range f.Fields {
		show(command, field, f.Responses[id])

		val, err := read(field, prompt)
		if err != nil {
			return err
		}

		res, err := save(field, val)
		if err != nil {
			return err
		}
		f.Responses[id] = res
	}

	return nil
}

func show(command *cobra.Command, field Field, savedValue interface{}) {
	command.Print(field.Prompt)
	if field.IsYesOrNo {
		command.Print(" (y/n)")
	}
	command.Print(": ")

	if savedValue != nil {
		command.Printf("(%v) ", savedValue)
	} else if field.DefaultValue != nil {
		command.Printf("(%v) ", field.DefaultValue)
	}
}

func read(field Field, prompt cmd.Prompt) (string, error) {
	var val string
	var err error

	if field.IsHidden {
		val, err = prompt.ReadPassword()
	} else {
		val, err = prompt.ReadString('\n')
	}
	if err != nil {
		return "", err
	}

	return strings.TrimSuffix(val, "\n"), nil
}

func save(field Field, val string) (interface{}, error) {
	if field.IsYesOrNo {
		switch strings.ToUpper(val) {
		case "Y", "YES":
			return true, nil
		case "N", "NO":
			return false, nil
		}
		return false, fmt.Errorf(errors.InvalidChoiceMsg, val)
	}

	if val == "" {
		return field.DefaultValue, nil
	}

	return val, nil
}
