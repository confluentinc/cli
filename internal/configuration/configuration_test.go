package configuration

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v3/pkg/config"
)

func TestGetConfigWhitelist(t *testing.T) {
	cfg := config.AuthenticatedCloudConfigMock()
	if err := cfg.Save(); err != nil {
		panic(err)
	}

	expected := map[string]*configFieldInfo{
		"disable_feature_flags": {name: "DisableFeatureFlags", kind: reflect.Bool},
		"disable_plugins":       {name: "DisablePlugins", kind: reflect.Bool},
		"disable_updates":       {name: "DisableUpdates", kind: reflect.Bool},
		"disable_update_check":  {name: "DisableUpdateCheck", kind: reflect.Bool},
		"no_browser":            {name: "NoBrowser", kind: reflect.Bool},
	}
	if runtime.GOOS == "windows" {
		expected["disable_plugins_once"] = &configFieldInfo{name: "DisablePluginsOnce", kind: reflect.Bool}
	}

	require.Equal(t, expected, getConfigWhitelist(cfg))
}
