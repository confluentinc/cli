package featureflags

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGetAnnouncementsOrDeprecation_BadFormat(t *testing.T) {
	require.Empty(t, GetAnnouncementsOrDeprecation(""))
}

func TestGetAnnouncementsOrDeprecation(t *testing.T) {
	resp := []any{
		map[string]any{
			"pattern": "command",
			"message": "0",
		},
		map[string]any{
			"pattern": "command --flag",
			"message": "1",
		},
		map[string]any{
			"pattern": "command-with-dashes",
			"message": "2",
		},
		map[string]any{
			"pattern": "--flag-only",
			"message": "3",
		},
		map[string]any{
			"pattern": "--multiple --flags",
			"message": "4",
		},
	}

	expected := map[string]*Messages{
		"command": {
			CommandMessage: "0",
			Flags:          []string{"flag"},
			FlagMessages:   []string{"1"},
		},
		"command-with-dashes": {
			CommandMessage: "2",
			Flags:          []string{},
			FlagMessages:   []string{},
		},
		"": {
			Flags:        []string{"flag-only", "multiple", "flags"},
			FlagMessages: []string{"3", "4", "4"},
		},
	}

	require.Equal(t, expected, GetAnnouncementsOrDeprecation(resp))
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
