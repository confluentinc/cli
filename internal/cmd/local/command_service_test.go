package local

import (
	"github.com/stretchr/testify/require"
	"strings"
	"testing"

	"github.com/confluentinc/cli/mock"
)

func TestServiceVersions(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(cp.AddFileToConfluentHome(file))

	versions := map[string]string{
		"Confluent Platform": "1.0.0",
		"kafka":              "2.0.0",
		"zookeeper":          "3.0.0",
	}

	for service, version := range versions {
		file := strings.Replace(versionFiles[service], "*", version, 1)
		req.NoError(cp.AddFileToConfluentHome(file))
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
