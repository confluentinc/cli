package connect

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/properties"
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

	err = updateWorkerConfig("new-plugin-dir", file.Name(), false)
	require.NoError(t, err)

	workerConfig, err := properties.LoadFile(file.Name(), properties.UTF8)
	require.NoError(t, err)

	require.Equal(t, workerConfig.Len(), 1)
	value, ok := workerConfig.Get("plugin.path")
	require.True(t, ok)
	require.Equal(t, value, "/usr/share/java, new-plugin-dir")
}
