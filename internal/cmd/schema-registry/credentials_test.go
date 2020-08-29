package schema_registry

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestSrAuthFound(t *testing.T) {
	req := require.New(t)

	cfg := mock.AuthenticatedDynamicConfigMock()
	cmd := &cobra.Command{}

	currCtx, err := cfg.Context(cmd)
	req.NoError(err)

	srCluster, err := currCtx.SchemaRegistryCluster(cmd)
	req.NoError(err)

	srAuth, didPromptUser, err := getSchemaRegistryAuth(cmd, srCluster.SrCredentials, false)
	req.NoError(err)

	req.False(didPromptUser)
	req.NotEmpty(srAuth.UserName)
	req.NotEmpty(srAuth.Password)
}
