package schemaregistry

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/require"

	climock "github.com/confluentinc/cli/mock"
)

func TestSrAuthFound(t *testing.T) {
	req := require.New(t)

	cfg := climock.AuthenticatedDynamicConfigMock()
	cmd := &cobra.Command{}

	ctx := cfg.Context()

	srCluster, err := ctx.SchemaRegistryCluster(cmd)
	req.NoError(err)

	srAuth, didPromptUser, err := getSchemaRegistryAuth(cmd, srCluster.SrCredentials, false)
	req.NoError(err)

	req.False(didPromptUser)
	req.NotEmpty(srAuth.UserName)
	req.NotEmpty(srAuth.Password)
}
