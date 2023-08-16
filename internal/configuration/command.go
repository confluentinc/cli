package configuration

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"

	"github.com/confluentinc/cli/v3/internal/update"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/types"
)

const fieldNotConfigurableError = `configuration field "%s" either does not exist or is not configurable`

type command struct {
	*pcmd.CLICommand
	cfg *config.Config
}

type configFieldInfo struct {
	kind     reflect.Kind
	name     string
	readOnly bool
}

type configFieldOut struct {
	Name     string `human:"Name" serialized:"name"`
	Value    string `human:"Value" serialized:"value"`
	ReadOnly bool   `human:"Read-Only" serialized:"read_only"`
}

func New(cfg *config.Config, prerunner pcmd.PreRunner) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "configuration",
		Aliases: []string{"config"},
		Short:   "Configure the Confluent CLI.",
	}

	c := &command{
		CLICommand: pcmd.NewAnonymousCLICommand(cmd, prerunner),
		cfg:        cfg,
	}

	cmd.AddCommand(c.newDescribeCommand())
	cmd.AddCommand(c.newListCommand())
	cmd.AddCommand(c.newUpdateCommand())

	return cmd
}

func getConfigWhitelist(cfg *config.Config) map[string]*configFieldInfo {
	whitelist := make(map[string]*configFieldInfo)
	t := reflect.TypeOf(*cfg)
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		// currently only boolean fields are part of this whitelist, but this may change in the future
		if kind := field.Type.Kind(); kind == reflect.Bool {
			if jsonFieldName := getJsonFieldName(field); jsonFieldName != "" {
				whitelist[jsonFieldName] = &configFieldInfo{
					kind:     kind,
					name:     field.Name,
					readOnly: isReadOnly(jsonFieldName),
				}
			}
		}
	}
	return whitelist
}

func getJsonFieldName(field reflect.StructField) string {
	jsonTag := field.Tag.Get("json")
	if jsonTag == "-" {
		return ""
	}
	if strings.Contains(jsonTag, ",") {
		jsonTag, _, _ = strings.Cut(jsonTag, ",")
	}
	// Want to hide this from linux and mac until the name specifies this field only affects Windows to avoid confusion
	if jsonTag == "disable_plugins_once" && runtime.GOOS != "windows" {
		return ""
	}
	return jsonTag
}

func isReadOnly(jsonField string) bool {
	return jsonField == "disable_updates" && update.IsHomebrew()
}

func (c *command) newConfigurationOut(field string, configWhitelist map[string]*configFieldInfo) *configFieldOut {
	value := reflect.ValueOf(c.cfg).Elem().FieldByName(configWhitelist[field].name)
	return &configFieldOut{
		Name:     field,
		Value:    fmt.Sprintf("%v", value),
		ReadOnly: configWhitelist[field].readOnly,
	}
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return types.GetSortedKeys(getConfigWhitelist(c.cfg))
}
