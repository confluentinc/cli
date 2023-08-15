package configuration

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/internal/cmd/update"
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/types"
)

const fieldNotConfigurableError = `config field "%s" either doesn't exist or is not configurable`

type command struct {
	*pcmd.CLICommand
	cfg             *config.Config
	jsonFieldToName map[string]string
	jsonFieldToType map[string]reflect.Kind
}

type configurationOut struct {
	Name  string `human:"Name" serialized:"name"`
	Value string `human:"Value" serialized:"value"`
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config"},
		Short:   "Configure the Confluent CLI.",
	}

	jsonFieldToType, jsonFieldToName := getSettableConfigFields(cfg)
	c := &command{
		CLICommand:      pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:             cfg,
		jsonFieldToName: jsonFieldToName,
		jsonFieldToType: jsonFieldToType,
	}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func getSettableConfigFields(cfg *config.Config) (map[string]reflect.Kind, map[string]string) {
	jsonFieldToType := make(map[string]reflect.Kind)
	jsonFieldToName := make(map[string]string)
	t := reflect.TypeOf(*cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if kind := field.Type.Kind(); kind == reflect.Bool {
			if jsonFieldName := getJsonFieldName(field, cfg.IsTest); jsonFieldName != "" {
				jsonFieldToType[jsonFieldName] = kind
				jsonFieldToName[jsonFieldName] = field.Name
			}
		}
	}
	return jsonFieldToType, jsonFieldToName
}

func getJsonFieldName(field reflect.StructField, isTest bool) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}
	if strings.Contains(jsonTag, ",") {
		jsonTag, _, _ = strings.Cut(jsonTag, ",")
	}
	if jsonTag == "disable_plugins_once" && runtime.GOOS != "windows" {
		return ""
	}
	if jsonTag == "disable_updates" && !isTest && update.IsHomebrew() {
		return ""
	}
	return jsonTag
}

func (c *command) newConfigurationOut(field string) *configurationOut {
	value := reflect.ValueOf(c.cfg).Elem().FieldByName(c.jsonFieldToName[field])
	configOut := &configurationOut{
		Name:  field,
		Value: fmt.Sprintf("%v", value),
	}
	return configOut
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return types.GetSortedKeys(c.jsonFieldToName)
}
