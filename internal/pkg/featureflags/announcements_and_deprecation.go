package featureflags

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	v1 "github.com/confluentinc/cli/internal/pkg/config/v1"
	dynamicconfig "github.com/confluentinc/cli/internal/pkg/dynamic-config"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

const deprecationPrefix = "DEPRECATED: "

const (
	Announcements      = "cli.announcements"
	DeprecationNotices = "cli.deprecation_notices"
)

type FlagsAndMsg struct {
	Flags        []string
	FlagMessages []string
	CmdMessage   string
}

func GetAnnouncementsOrDeprecation(ld interface{}) map[string]*FlagsAndMsg {
	cmdToFlagsAndMsg := make(map[string]*FlagsAndMsg)
	for _, val := range ld.([]interface{}) {
		if len(val.(map[string]interface{})) == 0 {
			break
		}
		var flags []string
		var msg = val.(map[string]interface{})["message"].(string)
		var command = val.(map[string]interface{})["pattern"].(string)
		if idx := strings.Index(command, "-"); idx != -1 {
			flags = strings.Split(command[idx:], " ")
			for i := range flags {
				flags[i] = strings.TrimLeft(flags[i], "-")
			}
			command = command[:idx-1]
		}
		if len(flags) == 0 {
			cmdToFlagsAndMsg[command] = &FlagsAndMsg{CmdMessage: msg}
		} else {
			msgs := make([]string, len(flags))
			for i := range msgs {
				msgs[i] = msg
			}
			if _, ok := cmdToFlagsAndMsg[command]; !ok {
				cmdToFlagsAndMsg[command] = &FlagsAndMsg{Flags: []string{}, FlagMessages: []string{}}
			}
			cmdToFlagsAndMsg[command].Flags = append(cmdToFlagsAndMsg[command].Flags, flags...)
			cmdToFlagsAndMsg[command].FlagMessages = append(cmdToFlagsAndMsg[command].FlagMessages, msgs...)
		}
	}
	return cmdToFlagsAndMsg
}

func DeprecateCommandTree(cmd *cobra.Command) {
	if cmd.Long != "" {
		cmd.Long = deprecationPrefix + cmd.Long
	}
	if cmd.Short != "" {
		cmd.Short = deprecationPrefix + cmd.Short
	}
	for _, subcommand := range cmd.Commands() {
		DeprecateCommandTree(subcommand)
	}
}

func DeprecateFlags(cmd *cobra.Command, flags []string) {
	for _, flag := range flags {
		if len(flag) == 1 {
			flag = cmd.Flags().ShorthandLookup(flag).Name
		}
		if cmd.Flag(flag) != nil {
			cmd.Flag(flag).Usage = deprecationPrefix + cmd.Flag(flag).Usage
		}
	}
	for _, subcommand := range cmd.Commands() {
		DeprecateFlags(subcommand, flags)
	}
}

func PrintAnnouncements(featureFlag string, ctx *dynamicconfig.DynamicContext, cmd *cobra.Command) {
	flagResponse := Manager.JsonVariation(featureFlag, ctx, v1.CliLaunchDarklyClient, true, []interface{}{})
	cmdToFlagsAndMsg := GetAnnouncementsOrDeprecation(flagResponse)
	for name, flagsAndMsg := range cmdToFlagsAndMsg {
		if strings.HasPrefix(cmd.CommandPath(), "confluent "+name) {
			if len(flagsAndMsg.Flags) == 0 {
				if featureFlag == DeprecationNotices {
					utils.ErrPrintln(cmd, fmt.Sprintf("`confluent %s` is deprecated: %s", name, flagsAndMsg.CmdMessage))
				} else {
					utils.ErrPrintln(cmd, flagsAndMsg.CmdMessage)
				}
			} else {
				for i, flag := range flagsAndMsg.Flags {
					var msg string
					if len(flag) == 1 {
						flag = cmd.Flags().ShorthandLookup(flag).Name
					}
					if cmd.Flags().Changed(flag) {
						if featureFlag == DeprecationNotices {
							msg = fmt.Sprintf("The `--%s` flag is deprecated", flag)
						}
						if flagsAndMsg.FlagMessages[i] == "" {
							msg = fmt.Sprintf("%s.", msg)
						} else {
							msg = fmt.Sprintf("%s: %s", msg, flagsAndMsg.FlagMessages[i])
						}
						utils.ErrPrintln(cmd, msg)
					}
				}
			}
		}
	}
}
