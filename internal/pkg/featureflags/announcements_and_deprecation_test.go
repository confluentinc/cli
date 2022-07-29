package featureflags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLDResponseToMap(t *testing.T) {
	ldResp := []map[string]interface{}{{"message": "DEPRECATED", "pattern": "ksql app"}, {"message": "DEPRECATED", "pattern": "kafka cluster list --all"}, {"message": "DEPRECATED", "pattern": "kafka cluster list --context"}}
	ld := make([]interface{}, len(ldResp))
	for i := range ldResp {
		ld[i] = ldResp[i]
	}
	cmdToFlagsAndMsg := LDResponseToMap(ld)
	var actual string
	for cmd, flagsAndMsg := range cmdToFlagsAndMsg {
		actual += fmt.Sprintln(cmd, flagsAndMsg.Flags, flagsAndMsg.FlagMessages, flagsAndMsg.CmdMessage)
	}
	require.Equal(t, "ksql app [] [] DEPRECATED\nkafka cluster list [--all --context] [DEPRECATED DEPRECATED] \n", actual)
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
	DeprecateFlags(cmds[0], []string{"--meaningless_flag"})
	for _, cmd := range cmds {
		require.Equal(t, "DEPRECATED: testing purposes", cmd.Flag("meaningless_flag").Usage)
	}
}

func dummyCmds() []*cobra.Command {
	cmds := make([]*cobra.Command, 4)
	for i := 0; i < 4; i++ {
		cmds[i] = &cobra.Command{
			Short: "short",
			Long:  "long description",
		}
		cmds[i].Flags().String("meaningless_flag", "true", "testing purposes")
	}
	cmds[0].AddCommand(cmds[1], cmds[2])
	cmds[1].AddCommand(cmds[3])
	return cmds
}
