package featureflags

import (
	"strings"

	"github.com/spf13/cobra"
	"golang.org/x/exp/slices"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	ppanic "github.com/confluentinc/cli/internal/pkg/panic-recovery"
)

func DisableHelpText(command *cobra.Command, flags []string) {
	if len(flags) == 0 {
		command.Hidden = true
	} else {
		formattedFlags := ppanic.ParseFlags(command, flags)
		for _, flag := range formattedFlags {
			_ = command.Flags().MarkHidden(flag)
		}
	}
}

func GetLDDisableMap(ctx *dynamicconfig.DynamicContext) map[string]any {
	ldDisableJson := Manager.JsonVariation("cli.disable", ctx, v1.CliLaunchDarklyClient, true, nil)
	ldDisable, ok := ldDisableJson.(map[string]any)
	if !ok {
		return nil
	}
	return ldDisable
}

func IsDisabled(cmd *cobra.Command, disabledPatterns []any) bool {
	for _, pattern := range disabledPatterns {
		if disabledCommand, args, err := cmd.Root().Find(strings.Split(pattern.(string), " ")); err == nil {
			trimmedFlags := ppanic.ParseFlags(disabledCommand, args)
			if len(trimmedFlags) == 0 {
				if strings.Contains(cmd.CommandPath(), disabledCommand.CommandPath()) {
					return true
				}
			} else {
				for _, disabledFlag := range trimmedFlags {
					if slices.Contains(Manager.flags, disabledFlag) {
						return true
					}
				}
			}
		}
	}
	return false
}
