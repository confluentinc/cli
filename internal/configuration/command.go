package configuration

import (
	"fmt"
	"reflect"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	"github.com/confluentinc/cli/v3/internal/update"
	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/types"
)

const fieldNotConfigurableError = `configuration field "%s" does not exist or is not configurable`

type command struct {
	*pcmd.CLICommand
	cfg *config.Config
}

type fieldInfo struct {
	kind     reflect.Kind
	name     string
	readOnly bool
}

type fieldOut struct {
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

func getWhitelist(cfg *config.Config) map[string]*fieldInfo {
	if runtime.GOOS == "windows" {
		config.Whitelist = append(config.Whitelist, "disable_plugins_once")
	}

	whitelist := make(map[string]*fieldInfo, len(config.Whitelist))
	t := reflect.TypeOf(*cfg)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag, _, _ := strings.Cut(field.Tag.Get("json"), ",")
		if slices.Contains(config.Whitelist, jsonTag) {
			whitelist[jsonTag] = &fieldInfo{
				kind:     field.Type.Kind(),
				name:     field.Name,
				readOnly: isReadOnly(jsonTag),
			}
		}
	}
	return whitelist
}

func isReadOnly(jsonField string) bool {
	return jsonField == "disable_updates" && update.IsHomebrew()
}

func (c *command) newFieldOut(field string, whitelist map[string]*fieldInfo) *fieldOut {
	value := reflect.ValueOf(c.cfg).Elem().FieldByName(whitelist[field].name)
	return &fieldOut{
		Name:     field,
		Value:    fmt.Sprintf("%v", value),
		ReadOnly: whitelist[field].readOnly,
	}
}

func (c *command) validArgs(cmd *cobra.Command, args []string) []string {
	if len(args) > 0 {
		return nil
	}

	if err := c.PersistentPreRunE(cmd, args); err != nil {
		return nil
	}

	return types.GetSortedKeys(getWhitelist(c.cfg))
}
