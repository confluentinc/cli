package configuration

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
	"github.com/confluentinc/cli/internal/pkg/types"
)

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

	cmd.AddCommand(c.newSetCommand())
	cmd.AddCommand(c.newViewCommand())

	return cmd
}

func getSettableConfigFields(cfg *config.Config) (map[string]reflect.Kind, map[string]string) {
	jsonFieldToType := make(map[string]reflect.Kind)
	jsonFieldToName := make(map[string]string)
	t := reflect.TypeOf(*cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if kind := field.Type.Kind(); kind == reflect.Bool {
			if jsonFieldName := getJsonFieldName(field); jsonFieldName != "" {
				jsonFieldToType[jsonFieldName] = kind
				jsonFieldToName[jsonFieldName] = field.Name
			}
		}
	}
	return jsonFieldToType, jsonFieldToName
}

func getJsonFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}
	if strings.Contains(jsonTag, ",") {
		jsonTag, _, _ = strings.Cut(jsonTag, ",")
	}
	if runtime.GOOS != "windows" && jsonTag == "disable_plugins_once" {
		return ""
	}
	return jsonTag
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
