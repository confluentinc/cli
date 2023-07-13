package plugin

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/hashicorp/go-version"
	"github.com/stretchr/testify/assert"

	"github.com/confluentinc/cli/internal/pkg/plugin"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func TestGetPluginManifest(t *testing.T) {
	dir, _ := filepath.Abs("../../../test/fixtures/input/plugin")
	manifest, err := getPluginManifest("confluent-test_plugin", dir)
	assert.NoError(t, err)

	referenceManifest := &Manifest{
		Name:        "confluent-test_plugin",
		Description: "Does nothing",
		Dependencies: []Dependency{
			{
				Name:    "Python",
				Version: "3",
			},
		},
	}
	assert.True(t, reflect.DeepEqual(referenceManifest, manifest))
}

func TestGetLanguage(t *testing.T) {
	dir, _ := filepath.Abs("../../../test/fixtures/input/plugin")
	manifest, err := getPluginManifest("confluent-test_plugin", dir)
	assert.NoError(t, err)

	language, ver, err := getLanguage(manifest)
	assert.NoError(t, err)
	assert.Equal(t, "python", language)
	referenceVer, err := version.NewVersion("3.0.0")
	assert.NoError(t, err)
	assert.True(t, ver.Equal(referenceVer))
}

func TestInstallPythonPlugin(t *testing.T) {
	dir, err := os.MkdirTemp("", "plugin-search")
	assert.NoError(t, err)
	defer os.RemoveAll(dir)

	pluginInstaller := &plugin.PythonPluginInstaller{
		Name:          "confluent-test_plugin",
		RepositoryDir: "../../../test/fixtures/input/plugin",
		InstallDir:    dir,
	}

	err = pluginInstaller.Install()
	assert.NoError(t, err)
	assert.True(t, utils.DoesPathExist(fmt.Sprintf("%s/%s", dir, "confluent-test_plugin.py")))
}
