package test_server

import (
	"testing"

	"github.com/gorilla/mux"
)

// v2 cmk urls
const (
	cmkCluster      = "/cmk/v2/clusters/{id}"
	cmkClusters     = "/cmk/v2/clusters"
	orgEnvironment  = "/org/v2/environments/{id}"
	orgEnvironments = "/org/v2/environments"
)

type V2Router struct {
	*mux.Router
	kafkaRPUrl string
}

func NewV2Router(t *testing.T) *V2Router {
	router := NewEmptyV2Router()
	router.buildV2Handler(t)
	return router
}

func NewEmptyV2Router() *V2Router {
	return &V2Router{
		Router: mux.NewRouter(),
	}
}

func (c *V2Router) buildV2Handler(t *testing.T) {
	c.addCmkClusterRoutes(t)
	c.addOrgEnvironmentRoutes(t)
}

func (c *V2Router) addCmkClusterRoutes(t *testing.T) {
	c.HandleFunc(cmkCluster, c.HandleCmkCluster(t))
	c.HandleFunc(cmkClusters, c.HandleCmkClusters(t))
}

func (c *V2Router) addOrgEnvironmentRoutes(t *testing.T) {
	c.HandleFunc(orgEnvironment, c.HandleOrgEnvironment(t))
	c.HandleFunc(orgEnvironments, c.HandleOrgEnvironments(t))
}
