package featureflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestDisableHelpText(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("flag1", "", "testing help text disable")
	cmd.Flags().String("flag2", "", "testing help text disable")
	cmd.Flags().StringP("flag3", "f", "", "testing help text disable")

	DisableHelpText(cmd, []string{})
	require.Equal(t, true, cmd.Hidden)

	DisableHelpText(cmd, []string{"--flag1", "-f", "--flag4"})
	require.Equal(t, true, cmd.Flag("flag1").Hidden)
	require.Equal(t, false, cmd.Flag("flag2").Hidden)
	require.Equal(t, true, cmd.Flag("flag3").Hidden)
}

func TestIsDisabled(t *testing.T) {
	cmds := disableCmds()

	require.Equal(t, true, IsDisabled(cmds[2], []any{"parent child"}))
	require.Equal(t, false, IsDisabled(cmds[2], []any{"parent other_child"}))
	require.Equal(t, true, IsDisabled(cmds[1], []any{"parent"}) &&
		IsDisabled(cmds[2], []any{"parent"}) && IsDisabled(cmds[3], []any{"parent"}))

	Manager.SetCommandAndFlags(cmds[0], []string{"parent", "child", "--flag1"})
	require.Equal(t, true, IsDisabled(cmds[2], []any{"parent child --flag1"}))

	Manager.SetCommandAndFlags(cmds[0], []string{"parent", "child"})
	require.Equal(t, false, IsDisabled(cmds[2], []any{"parent child --flag1"}))

	Manager.SetCommandAndFlags(cmds[0], []string{"parent", "other_child", "-f"})
	require.Equal(t, true, IsDisabled(cmds[3], []any{"parent other_child -f"}))
}

func disableCmds() []*cobra.Command {
	cmds := make([]*cobra.Command, 4)
	for i := 0; i < 4; i++ {
		cmds[i] = &cobra.Command{}
	}
	cmds[0].AddCommand(cmds[1])

	cmds[1].Use = "parent"
	cmds[1].AddCommand(cmds[2])
	cmds[1].AddCommand(cmds[3])

	cmds[2].Use = "child"
	cmds[2].Flags().String("flag1", "true", "testing flag disable")

	cmds[3].Use = "other_child"
	cmds[3].Flags().StringP("flag2", "f", "true", "testing shorthand disable")
	return cmds
}
