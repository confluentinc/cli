package featureflags

import (
	"strings"

	"github.com/spf13/cobra"
)

type FlagsAndMsg struct {
	Flags   []string
	Message string
}

func LDResponseToMap(ld interface{}) map[string]FlagsAndMsg {
	cmdToFlagsAndMsg := make(map[string]FlagsAndMsg)
	for _, val := range ld.([]interface{}) {
		flags := ""
		var flagNames []string
		var msg = val.(map[string]interface{})["message"].(string)
		var command = val.(map[string]interface{})["pattern"].(string)
		if strings.Contains(command, "--") {
			flags = command[strings.Index(command, "--"):]
			flagNames = strings.Split(flags, " ")
			for i, flag := range flagNames {
				flagNames[i] = strings.TrimPrefix(flag, "--")
			}
			command = command[:strings.Index(command, "--")]
		}
		cmdToFlagsAndMsg[command] = FlagsAndMsg{Flags: flagNames, Message: msg}
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
