package configuration

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"
	"golang.org/x/exp/slices"

	"github.com/confluentinc/cli/v3/pkg/config"
	"github.com/confluentinc/cli/v3/pkg/types"
)

func TestGetConfigWhitelist(t *testing.T) {
	cfg := config.AuthenticatedCloudConfigMock()
	if err := cfg.Save(); err != nil {
		panic(err)
	}

	expected := []string{
		"disable_feature_flags",
		"disable_plugins",
		"disable_update_check",
		"disable_updates",
		"no_browser",
	}

	if runtime.GOOS == "windows" {
		expected = append(expected, "disable_plugins_once")
		slices.Sort(expected)
	}

	require.Equal(t, expected, types.GetSortedKeys(getWhitelist(cfg)))
}
