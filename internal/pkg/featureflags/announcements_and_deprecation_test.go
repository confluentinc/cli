package featureflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetAnnouncementsOrDeprecation(t *testing.T) {
	ldResp := []map[string]interface{}{
		{"message": "DEPRECATED", "pattern": "kafka cluster list --all --context"},
	}
	ld := make([]interface{}, len(ldResp))
	for i := range ldResp {
		ld[i] = ldResp[i]
	}
	cmdToFlagsAndMsg := GetAnnouncementsOrDeprecation(ld)
	expected := map[string]*FlagsAndMsg{}
	expected["kafka cluster list"] = &FlagsAndMsg{
		Flags:        []string{"all", "context"},
		FlagMessages: []string{"DEPRECATED", "DEPRECATED"},
	}
	for cmd, flagsAndMsg := range cmdToFlagsAndMsg {
		require.Equal(t, expected[cmd], flagsAndMsg)
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
	DeprecateFlags(cmds[0], []string{"flag1", "f"})
	for _, cmd := range cmds {
		require.Equal(t, "DEPRECATED: testing flag deprecation", cmd.Flag("flag1").Usage)
		require.Equal(t, "DEPRECATED: testing shorthand deprecation", cmd.Flag("flag2").Usage)
	}
}

func dummyCmds() []*cobra.Command {
	cmds := make([]*cobra.Command, 3)
	for i := 0; i < 3; i++ {
		cmds[i] = &cobra.Command{
			Short: "short",
			Long:  "long description",
		}
		cmds[i].Flags().String("flag1", "true", "testing flag deprecation")
		cmds[i].Flags().StringP("flag2", "f", "true", "testing shorthand deprecation")
	}
	cmds[0].AddCommand(cmds[1])
	cmds[1].AddCommand(cmds[2])
	return cmds
}
