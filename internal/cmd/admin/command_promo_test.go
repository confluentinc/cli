package admin

import (
	pcmd "github.com/confluentinc/cli/internal/pkg/cmd"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestPromoAdd(t *testing.T) {
	cmd := mockAdminCommand()

	out, err := pcmd.ExecuteCommand(cmd, "promo", "add", "XXXXX")
	require.NoError(t, err)
	require.Equal(t, "Your promo code was successfully added.\n", out)
}
