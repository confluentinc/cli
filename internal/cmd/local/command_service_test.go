package local

import (
	"io/ioutil"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestConfigService(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())
	req.NoError(cp.AddFileToConfluentHome("etc/kafka/zookeeper.properties", "replace=old\n# comment=old\n", 0644))
	req.NoError(cp.NewConfluentCurrent())

	dir, err := getServiceDir("zookeeper")
	req.NoError(err)
	config := map[string]string{"replace": "new", "comment": "new", "append": "new"}
	req.NoError(configService("zookeeper", dir, config))

	properties := filepath.Join(cp.ConfluentCurrent, "zookeeper", "zookeeper.properties")
	req.FileExists(properties)
	data, err := ioutil.ReadFile(properties)
	req.NoError(err)
	req.NotContains(string(data), "replace=old")
	req.Contains(string(data), "replace=new")
	req.NotContains(string(data), "# comment=old")
	req.Contains(string(data), "comment=new")
	req.Contains(string(data), "append=new")
}

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentHome())

	versions := map[string]string{
		"Confluent Platform": "1.0.0",
		"kafka":              "2.0.0",
		"zookeeper":          "3.0.0",
	}

	for service, version := range versions {
		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(cp.AddEmptyFileToConfluentHome(file))
	}

	for service := range services {
		out, err := mockLocalCommand("services", service, "version")
		req.NoError(err)

		version, ok := versions[service]
		if !ok {
			version = versions["Confluent Platform"]
		}
		req.Contains(out, version)
	}
}
