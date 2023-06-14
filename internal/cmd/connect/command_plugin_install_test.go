package connect

import (
	"archive/zip"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/cpstructs"
	"github.com/confluentinc/cli/internal/pkg/utils"
)

type MockCommand struct {
	OutputStr string
}

func (c *MockCommand) Output() ([]byte, error) {
	return []byte(c.OutputStr), nil
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

	manifest := &cpstructs.Manifest{
		Name:    "unit-test-plugin",
		Version: "0.0.0",
		Owner: cpstructs.Owner{Username: "confluentinc"},
	}

	zipReader, err := zip.OpenReader(archivePath)
	require.NoError(t, err)
	defer zipReader.Close()

	err = unzipPlugin(manifest, zipReader.File, tempDir)
	require.NoError(t, err)
	require.True(t, utils.DoesPathExist(fmt.Sprintf("%s/confluentinc-unit-test-plugin/test-plugin/manifest.json", tempDir)))
}

func TestRunningWorkerConfigLocations(t *testing.T) {
	outputStr := "12345 s006  S      1:08.51 java -Xms256M -Xmx2G -server -XX:+UseG1GC " +
		"-XX:MaxGCPauseMillis=20 -XX:InitiatingHeapOccupancyPercent=35 " +
		"-XX:+ExplicitGCInvokesConcurrent -Djava.awt.headless=true " +
		"-Dcom.sun.management.jmxremote -Dcom.sun.management.jmxremote.authenticate=false " +
		"-Dcom.sun.management.jmxremote.ssl=false " +
		"-Dkafka.logs.dir=/var/folders/00/XXX/T/confluent.abc123/connect/logs " +
		"-cp " +
		"/Users/namegoeshere/confluent-platform/share/java/kafka/*: " +
		"/Users/namegoeshere/confluent-platform/share/java/confluent-common/*: " +
		"/Users/namegoeshere/confluent-platform/share/java/kafka-serde-tools/*: " +
		"/Users/namegoeshere/confluent-platform/share/java/monitoring-interceptors/*: " +
		"/Users/namegoeshere/confluent-platform/bin/../share/java/kafka/*: " +
		"/Users/namegoeshere/confluent-platform/bin/../share/java/confluent-support-metrics/*: " +
		"/usr/share/java/confluent-support-metrics/* org.apache.kafka.connect.cli.ConnectDistributed " +
		"-daemon " +
		"/var/folders/00/XXX/T/confluent.abc123/connect/connect.properties " +
		"/Users/namegoeshere/artifacts/confluent-5.0.0-SNAPSHOT/etc/kafka-connect-replicator/replicator-connect-standalone.properties"
	searchProcessCmd := &MockCommand{OutputStr: outputStr}

	runningWorkerConfigs, err := runningWorkerConfigLocations(searchProcessCmd)
	require.NoError(t, err)

	expectedConfigs := []WorkerConfig{
		{
			Path: "/var/folders/00/XXX/T/confluent.abc123/connect/connect.properties",
			Use:  "Used by Connect process with PID 12345",
		},
	}
	require.Equal(t, runningWorkerConfigs, expectedConfigs)
}
