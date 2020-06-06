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
	req.Equal(orderedServices, availableServices)
}

func TestAvailableServicesNoConfluentPlatform(t *testing.T) {
	req := require.New(t)

	cp := mock.NewConfluentPlatform()
	defer cp.TearDown()

	req.NoError(cp.NewConfluentHome())

	servicesNoControlCenter := []string{
		"zookeeper",
		"kafka",
		"connect",
		"kafka-rest",
		"schema-registry",
		"ksql-server",
	}
	availableServices, err := getAvailableServices()
	req.NoError(err)
	req.Equal(servicesNoControlCenter, availableServices)

}
