package configuration

import (
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/confluentinc/cli/internal/pkg/config"
)

type command struct {
	*pcmd.CLICommand
	config *config.Config
}

type configurationOut struct {
	Name     string `human:"Name" serialized:"name"`
	Value    string `human:"Value" serialized:"value"`
	ReadOnly bool   `human:"Read-Only" serialized:"read_only"`
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config"},
		Short:   "Manage CLI configuration fields.",
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		config:     cfg,
	}

	cmd.AddCommand(c.newSetCommand())
	cmd.AddCommand(c.newListCommand())

	return cmd
}

func getSettableConfigFields(cfg *config.Config) (map[string]reflect.Kind, map[string]string) {
	jsonFieldToType := make(map[string]reflect.Kind)
	jsonFieldToName := make(map[string]string)
	t := reflect.TypeOf(*cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		kind := field.Type.Kind()
		switch kind {
		case reflect.Bool, reflect.String:
			jsonFieldName := getJsonFieldName(field)
			if jsonFieldName != "" {
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
