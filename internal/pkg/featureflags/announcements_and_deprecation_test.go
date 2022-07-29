package featureflags

import (
	"fmt"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestLDResponseToMap(t *testing.T) {
	ldResp := []map[string]interface{}{{"message": "DEPRECATED", "pattern": "ksql app"}, {"message": "DEPRECATED", "pattern": "kafka cluster list --all"}}
	ld := make([]interface{}, len(ldResp))
	for i := range ldResp {
		ld[i] = ldResp[i]
	}
	cmdToFlagsAndMsg := LDResponseToMap(ld)
	require.Equal(t, "map[kafka cluster list :{[all] DEPRECATED} ksql app:{[] DEPRECATED}]", fmt.Sprint(cmdToFlagsAndMsg))
}

func TestDeprecateCommandTree(t *testing.T) {
	cmds := make([]*cobra.Command, 4)
	for i := 0; i < 4; i++ {
		cmds[i] = &cobra.Command{
			Short: "short",
			Long:  "long description",
		}
	}
	cmds[0].AddCommand(cmds[1], cmds[2])
	cmds[1].AddCommand(cmds[3])
	DeprecateCommandTree(cmds[0])
	for _, cmd := range cmds {
		require.Equal(t, "DEPRECATED: short", cmd.Short)
		require.Equal(t, "DEPRECATED: long description", cmd.Long)
	}
}
