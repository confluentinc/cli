package connect

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

func TestCompactDuplicateInstallations(t *testing.T) {
	installations := []platformInstallation{
		{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: "/path/1",
			},
			Use: "$CONFLUENT_HOME",
		},
		{
			Location: platformLocation{
				Type: "PACKAGE",
				Path: "/path/2",
			},
			Use: "Installed RPM/DEB Package",
		},
		{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: "/path/1",
			},
			Use: "Archive Duplicate",
		},
		{
			Location: platformLocation{
				Type: "PACKAGE",
				Path: "/path/2",
			},
			Use: "Package Duplicate",
		},
		{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: "/path/3",
			},
			Use: "Another Archive",
		},
	}

	comparison := []platformInstallation{
		{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: "/path/1",
			},
			Use: "$CONFLUENT_HOME",
		},
		{
			Location: platformLocation{
				Type: "PACKAGE",
				Path: "/path/2",
			},
			Use: "Installed RPM/DEB Package",
		},
		{
			Location: platformLocation{
				Type: "ARCHIVE",
				Path: "/path/3",
			},
			Use: "Another Archive",
		},
	}

	compactedInstallations := compactDuplicateInstallations(installations)

	require.Equal(t, compactedInstallations, comparison)
}

func TestUpdateWorkerConfig(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "worker-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	file, err := os.Create(fmt.Sprintf("%s/test.properties", tempDir))
	require.NoError(t, err)
	_, err = file.Write([]byte("plugin.path=/usr/share/java"))
	require.NoError(t, err)

	// Dry run: expect no changes
	err = updateWorkerConfig("new-plugin-dir", file.Name(), true)
	require.NoError(t, err)

	workerConfig, err := properties.LoadFile(file.Name(), properties.UTF8)
	require.NoError(t, err)

	require.Equal(t, workerConfig.Len(), 1)
	value, ok := workerConfig.Get("plugin.path")
	require.True(t, ok)
	require.Equal(t, value, "/usr/share/java")

	// Actual run
	err = updateWorkerConfig("new-plugin-dir", file.Name(), false)
	require.NoError(t, err)

	workerConfig, err = properties.LoadFile(file.Name(), properties.UTF8)
	require.NoError(t, err)

	require.Equal(t, workerConfig.Len(), 1)
	value, ok = workerConfig.Get("plugin.path")
	require.True(t, ok)
	require.Equal(t, value, "/usr/share/java, new-plugin-dir")
}

func TestUnzipPlugin(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "zip-test")
	require.NoError(t, err)
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	archivePath := fmt.Sprintf("%s/test-plugin.zip", tempDir)
	err = newTestZip(archivePath)
	require.NoError(t, err)

	manifest := &ccstructs.Manifest{
		Name:    "unit-test-plugin",
		Version: "0.0.0",
		Owner: ccstructs.Owner{
			Username: "confluentinc",
		},
	}

	zipReader, err := zip.OpenReader(archivePath)
	require.NoError(t, err)
	defer zipReader.Close()

	unzipPlugin(manifest, zipReader.File, tempDir)
	require.True(t, utils.DoesPathExist(fmt.Sprintf("%s/confluentinc-unit-test-plugin/test-plugin/manifest.json", tempDir)))
}

func newTestZip(archivePath string) error {
	zipFile, err := os.Create(archivePath)
	if err != nil {
		return err
	}
	defer zipFile.Close()
	writer := zip.NewWriter(zipFile)
	defer writer.Close()
	_, err = writer.Create("test-plugin/manifest.json")
	return err
}
