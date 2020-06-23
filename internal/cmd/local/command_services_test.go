package local

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestConfluentPlatformAvailableServices(t *testing.T) {
	req := require.New(t)

	c := &LocalCommand{
		ch: &mock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return true, nil
			},
		},
	}

	got, err := c.getAvailableServices()
	req.NoError(err)

	want := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
		"control-center",
	}
	req.Equal(want, got)
}

func TestConfluentCommunitySoftwareAvailableServices(t *testing.T) {
	req := require.New(t)

	c := &LocalCommand{
		ch: &mock.MockConfluentHome{
			IsConfluentPlatformFunc: func() (bool, error) {
				return false, nil
			},
		},
	}

	got, err := c.getAvailableServices()
	req.NoError(err)

	want := []string{
		"zookeeper",
		"kafka",
		"schema-registry",
		"kafka-rest",
		"connect",
		"ksql-server",
	}
	req.Equal(want, got)
}
