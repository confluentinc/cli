package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

const exampleVersion = "0.0.0"

func TestGetVersion(t *testing.T) {
	req := require.New(t)

	ch := &mock.MockConfluentHome{
		FindFileFunc: func(pattern string) ([]string, error) {
			versionFile := strings.ReplaceAll(versionFiles[exampleService], "*", exampleVersion)
			return []string{versionFile}, nil
		},
	}

	version, err := getVersion(ch, exampleService)
	req.NoError(err)
	req.Equal(exampleVersion, version)
}

func TestGetVersionNoMatchError(t *testing.T) {
	req := require.New(t)

	ch := &mock.MockConfluentHome{
		FindFileFunc: func(pattern string) ([]string, error) {
			return []string{}, nil
		},
	}

	_, err := getVersion(ch, exampleService)
	req.Error(err)
}
