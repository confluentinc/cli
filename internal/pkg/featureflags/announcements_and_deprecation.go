package featureflags

import (
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const deprecate = "DEPRECATED: "

type FlagsAndMsg struct {
	Flags        []string
	FlagMessages []string
	CmdMessage   string
}

func GetAnnouncementsAndDeprecation(ld interface{}) map[string]*FlagsAndMsg {
	cmdToFlagsAndMsg := make(map[string]*FlagsAndMsg)
	for _, val := range ld.([]interface{}) {
		var flag string
		var msg = val.(map[string]interface{})["message"].(string)
		var command = val.(map[string]interface{})["pattern"].(string)
		if ind := strings.Index(command, "--"); ind != -1 {
			flag = command[ind+2:]
			command = command[:ind-1]
		}
		if flag == "" {
			cmdToFlagsAndMsg[command] = &FlagsAndMsg{CmdMessage: msg}
		} else {
			if flagsAndMsg, ok := cmdToFlagsAndMsg[command]; ok {
				flagsAndMsg.Flags = append(flagsAndMsg.Flags, flag)
				flagsAndMsg.FlagMessages = append(flagsAndMsg.FlagMessages, msg)
			} else {
				cmdToFlagsAndMsg[command] = &FlagsAndMsg{Flags: []string{flag}, FlagMessages: []string{msg}}
			}
		}
	}
	return cmdToFlagsAndMsg
}

func DeprecateCommandTree(cmd *cobra.Command) {
	cmd.Short = deprecate + cmd.Short
	cmd.Long = deprecate + cmd.Long
	for _, subcommand := range cmd.Commands() {
		DeprecateCommandTree(subcommand)
	}
}

func DeprecateFlags(cmd *cobra.Command, flags []string) {
	for _, flag := range flags {
		if cmd.Flag(flag) != nil {
			cmd.Flag(flag).Usage = deprecate + cmd.Flag(flag).Usage
		}
	}
	for _, subcommand := range cmd.Commands() {
		DeprecateFlags(subcommand, flags)
	}
}

func PrintCmdMessages(featureFlag string, ctx *dynamicconfig.DynamicContext, cmd *cobra.Command) {
	flagResponse := Manager.JsonVariation(featureFlag, ctx, v1.CliLaunchDarklyClient, true, []interface{}{})
	cmdToFlagsAndMsg := GetAnnouncementsAndDeprecation(flagResponse)
	for name, flagsAndMsg := range cmdToFlagsAndMsg {
		if strings.HasPrefix(cmd.CommandPath(), "confluent "+name) {
			if len(flagsAndMsg.Flags) == 0 {
				utils.ErrPrintln(cmd, flagsAndMsg.CmdMessage)
			} else {
				for i, flag := range flagsAndMsg.Flags {
					if cmd.Flags().Changed(flag) {
						utils.ErrPrintln(cmd, flagsAndMsg.FlagMessages[i])
					}
				}
			}
		}
	}
}
