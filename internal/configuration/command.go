package configuration

import (
	"fmt"
	"reflect"
	"runtime"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	pcmd "github.com/confluentinc/cli/v3/pkg/cmd"
	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/types"
)

const (
	fieldDoesNotExistError = `configuration key "%s" does not exist`
	fieldReadOnlyError     = `configuration "%s" is read-only`
)

type command struct {
	*pcmd.CLICommand
	cfg *config.Config
}

type fieldInfo struct {
	kind reflect.Kind
	name string
}

type fieldOut struct {
	Name  string `human:"Name" serialized:"name"`
	Value string `human:"Value" serialized:"value"`
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
	fields := []string{
		"disable_feature_flags",
		"disable_plugins",
		"enable_color",
	}
	if runtime.GOOS == "windows" {
		fields = append(fields, "disable_plugins_once_windows")
	}
	if !cfg.DisableUpdates {
		fields = append(fields, "disable_update_check")
	}

	whitelist := make(map[string]*fieldInfo, len(fields))
	t := reflect.TypeOf(*cfg)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		jsonTag := strings.Split(field.Tag.Get("json"), ",")[0]
		if slices.Contains(fields, jsonTag) {
			whitelist[jsonTag] = &fieldInfo{
				kind: field.Type.Kind(),
				name: field.Name,
			}
		}
	}

	return whitelist
}

func (c *command) newFieldOut(field string, whitelist map[string]*fieldInfo) *fieldOut {
	value := reflect.ValueOf(c.cfg).Elem().FieldByName(whitelist[field].name)
	return &fieldOut{
		Name:  field,
		Value: fmt.Sprintf("%v", value),
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
