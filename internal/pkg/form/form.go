package form

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/pkg/cmd"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
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
	Fields    []Field
	Responses map[string]interface{}
}

type Field struct {
	ID           string
	Prompt       string
	DefaultValue interface{}
	IsYesOrNo    bool
	IsHidden     bool
}

func New(fields ...Field) *Form {
	return &Form{
		Fields:    fields,
		Responses: make(map[string]interface{}),
	}
}

func (f *Form) Prompt(command *cobra.Command, prompt cmd.Prompt) error {
	for _, field := range f.Fields {
		show(command, field, f.Responses[field.ID])

		val, err := read(field, prompt)
		if err != nil {
			return err
		}

		res, err := save(field, val)
		if err != nil {
			return err
		}
		f.Responses[field.ID] = res
	}

	return nil
}

func show(cmd *cobra.Command, field Field, savedValue interface{}) {
	pcmd.Print(cmd, field.Prompt)
	if field.IsYesOrNo {
		pcmd.Print(cmd, " (y/n)")
	}
	pcmd.Print(cmd, ": ")

	if savedValue != nil {
		pcmd.Printf(cmd, "(%v) ", savedValue)
	} else if field.DefaultValue != nil {
		pcmd.Printf(cmd, "(%v) ", field.DefaultValue)
	}
}

func read(field Field, prompt cmd.Prompt) (string, error) {
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

	if val == "" && field.DefaultValue != nil {
		return field.DefaultValue, nil
	}

	return val, nil
}
