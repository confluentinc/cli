package configuration

import (
	"fmt"
	"reflect"
	"strconv"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/errors"
	"github.com/confluentinc/cli/v3/pkg/examples"
	"github.com/confluentinc/cli/v3/pkg/form"
	"github.com/confluentinc/cli/v3/pkg/output"
)

func (c *command) newUpdateCommand() *cobra.Command {
	return &cobra.Command{
		Use:               "update [key] [value]",
		Short:             "Update a user-configurable field's value.",
		Args:              cobra.ExactArgs(2),
		ValidArgsFunction: pcmd.NewValidArgsFunction(c.validArgs),
		RunE:              c.update,
		Example: examples.BuildExampleString(
			examples.Example{
				Text: `Disable plugins by setting "disable_plugins" to true.`,
				Code: "confluent configuration update disable_plugins true",
			},
		),
	}
}

func (c *command) update(_ *cobra.Command, args []string) error {
	field := args[0]
	value, err := c.convertValue(field, args[1])
	if err != nil {
		if field == "current_context" {
			return errors.NewErrorWithSuggestions(err.Error(), "Please use `confluent context use` to set the current context.")
		}
		return err
	}

	oldValue := reflect.ValueOf(c.cfg).Elem().FieldByName(c.configWhiteList[field].name)
	if field == "disable_feature_flags" {
		if ok, err := confirmSet(field, form.NewPrompt()); err != nil {
			return err
		} else if !ok {
			return nil
		}
	}
	oldValue.Set(reflect.ValueOf(value))
	if err := c.cfg.Validate(); err != nil {
		return err
	}
	if err := c.cfg.Save(); err != nil {
		return err
	}
	output.Print(fmt.Sprintf(errors.UpdateSuccessMsg, "value", "config field", field, value))
	return nil
}

func (c *command) convertValue(field, value string) (any, error) {
	fieldInfo, ok := c.configWhiteList[field]
	if !ok {
		return nil, fmt.Errorf(fieldNotConfigurableError, field)
	}
	switch fieldInfo.kind {
	case reflect.Bool:
		val, err := strconv.ParseBool(value)
		if err != nil {
			return nil, fmt.Errorf(`"%s" is not a valid value for config field "%s", which is of type: %s`, value, field, fieldInfo.kind.String())
		}
		return val, nil
	default:
		return value, nil
	}
}

func confirmSet(field string, prompt form.Prompt) (bool, error) {
	f := form.New(
		form.Field{
			ID:        "proceed",
			Prompt:    fmt.Sprintf(`We don't recommend modifying the value of "%s", would you like to proceed?`, field),
			IsYesOrNo: true,
		},
	)
	if err := f.Prompt(prompt); err != nil {
		return false, err
	}
	if !f.Responses["proceed"].(bool) {
		output.Println(fmt.Sprintf(`Configuration field "%s" was not updated.`, field))
		return false, nil
	}
	return true, nil
}
