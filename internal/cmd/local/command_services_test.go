package local

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/mock"
)

func TestConfluentPlatformAvailableServices(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentHome())

	availableServices, err := getAvailableServices()
	req.NoError(err)
	req.Equal(orderedServices, availableServices)
}

func TestConfluentCommunitySoftwareAvailableServices(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentCommunitySoftware()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentHome())

	availableServices, err := getAvailableServices()
	req.NoError(err)
	req.NotContains(availableServices, "control-center")
}

func TestTopErrorNoRunningServices(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()
	req.NoError(cp.NewConfluentCurrent())

	_, err := mockLocalCommand("services", "top")
	req.Error(err)
}
