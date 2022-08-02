package featureflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetAnnouncementsAndDeprecation(t *testing.T) {
	ldResp := []map[string]interface{}{
		{"message": "DEPRECATED", "pattern": "ksql app"},
		{"message": "DEPRECATED", "pattern": "kafka cluster list --all"},
		{"message": "DEPRECATED", "pattern": "kafka cluster list --context"},
	}
	ld := make([]interface{}, len(ldResp))
	for i := range ldResp {
		ld[i] = ldResp[i]
	}
	cmdToFlagsAndMsg := GetAnnouncementsAndDeprecation(ld)
	expected := map[string]*FlagsAndMsg{}
	expected["ksql app"] = &FlagsAndMsg{CmdMessage: "DEPRECATED"}
	expected["kafka cluster list"] = &FlagsAndMsg{
		Flags:        []string{"all", "context"},
		FlagMessages: []string{"DEPRECATED", "DEPRECATED"},
	}
	for cmd, flagsAndMsg := range cmdToFlagsAndMsg {
		require.Equal(t, flagsAndMsg.Flags, expected[cmd].Flags)
		require.Equal(t, flagsAndMsg.FlagMessages, expected[cmd].FlagMessages)
		require.Equal(t, flagsAndMsg.CmdMessage, expected[cmd].CmdMessage)
	}
}

func TestDeprecateCommandTree(t *testing.T) {
	cmds := dummyCmds()
	DeprecateCommandTree(cmds[0])
	for _, cmd := range cmds {
		require.Equal(t, "DEPRECATED: short", cmd.Short)
		require.Equal(t, "DEPRECATED: long description", cmd.Long)
	}
}

func TestDeprecateFlags(t *testing.T) {
	cmds := dummyCmds()
	DeprecateFlags(cmds[0], []string{"meaningless-flag"})
	for _, cmd := range cmds {
		require.Equal(t, "DEPRECATED: testing purposes", cmd.Flag("meaningless-flag").Usage)
	}
}

func dummyCmds() []*cobra.Command {
	cmds := make([]*cobra.Command, 3)
	for i := 0; i < 3; i++ {
		cmds[i] = &cobra.Command{
			Short: "short",
			Long:  "long description",
		}
		cmds[i].Flags().String("meaningless-flag", "true", "testing purposes")
	}
	cmds[0].AddCommand(cmds[1])
	cmds[1].AddCommand(cmds[2])
	return cmds
}
