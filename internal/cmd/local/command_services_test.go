package local

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

const exampleDir = "dir"

func TestGetDataDirConfig(t *testing.T) {
	req := require.New(t)

	cc := &mock.MockConfluentCurrent{
		GetDataDirFunc: func(service string) (string, error) {
			return exampleDir, nil
		},
	}
	config, err := getDataDirConfig(cc, exampleService)

	req.NoError(err)
	req.Len(config, 1)
}

func TestConfluentPlatformAvailableServices(t *testing.T) {
	req := require.New(t)

	ch := &mock.MockConfluentHome{
		IsConfluentPlatformFunc: func() (bool, error) {
			return true, nil
		},
	}

	availableServices, err := getAvailableServices(ch)
	req.NoError(err)

	services := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
		"control-center",
	}
	req.Equal(services, availableServices)
}

func TestConfluentCommunitySoftwareAvailableServices(t *testing.T) {
	req := require.New(t)

	ch := &mock.MockConfluentHome{
		IsConfluentPlatformFunc: func() (bool, error) {
			return false, nil
		},
	}

	availableServices, err := getAvailableServices(ch)
	req.NoError(err)

	services := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
	}
	req.Equal(services, availableServices)
}
