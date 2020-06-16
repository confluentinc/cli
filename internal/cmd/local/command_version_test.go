package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestConfluentCommunitySoftwareVersion(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentCommunitySoftware()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	file := strings.Replace(versionFiles["Confluent Community Software"], "*", "0.0.0", 1)
	req.NoError(cp.AddEmptyFileToConfluentHome(file))

	out, err := mockLocalCommand("version")
	req.NoError(err)
	req.Contains(out, "Confluent Community Software: 0.0.0")
}

func TestConfluentPlatformVersion(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	file := strings.Replace(versionFiles["Confluent Platform"], "*", "1.0.0", 1)
	req.NoError(cp.AddEmptyFileToConfluentHome(file))

	out, err := mockLocalCommand("version")
	req.NoError(err)
	req.Contains(out, "Confluent Platform: 1.0.0")
}
