package featureflags

import (
	"fmt"
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
	require.Equal(t, fmt.Sprint(cmdToFlagsAndMsg), "map[kafka cluster list :{[all] DEPRECATED} ksql app:{[] DEPRECATED}]")
}
