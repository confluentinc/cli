package local

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestConfluentPlatformAvailableServices(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	file := strings.Replace(confluentControlCenter, "*", "0.0.0", 1)
	req.NoError(cp.AddFileToConfluentHome(file))

	availableServices, err := getAvailableServices()
	req.NoError(err)
	req.Equal(availableServices, topologicallySortedServices)
}

func TestAvailableServicesNoConfluentPlatform(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	availableServices, err := getAvailableServices()
	req.NoError(err)
	req.Equal(availableServices, topologicallySortedServices[:len(topologicallySortedServices)-1])
}
