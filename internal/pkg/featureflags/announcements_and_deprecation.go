package featureflags

import (
	"strings"

	"github.com/spf13/cobra"
)

type FlagsAndMsg struct {
	Flags        []string
	FlagMessages []string
	CmdMessage   string
}

func LDResponseToMap(ld interface{}) map[string]*FlagsAndMsg {
	cmdToFlagsAndMsg := make(map[string]*FlagsAndMsg)
	for _, val := range ld.([]interface{}) {
		var flag string
		var msg = val.(map[string]interface{})["message"].(string)
		var command = val.(map[string]interface{})["pattern"].(string)
		if strings.Contains(command, "--") {
			flag = command[strings.Index(command, "--")+2:]
			command = command[:strings.Index(command, "--")-1]
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
	cmd.Short = "DEPRECATED: " + cmd.Short
	cmd.Long = "DEPRECATED: " + cmd.Long
	for _, subcommand := range cmd.Commands() {
		DeprecateCommandTree(subcommand)
	}
}

func DeprecateFlags(cmd *cobra.Command, flags []string) {
	for _, flag := range flags {
		if cmd.Flag(flag) != nil {
			cmd.Flag(flag).Usage = "DEPRECATED: " + cmd.Flag(flag).Usage
		}
	}
	for _, subcommand := range cmd.Commands() {
		DeprecateFlags(subcommand, flags)
	}
}
