package test_server

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

// v2 cmk/org urls
const (
	cmkCluster      = "/cmk/v2/clusters/{id}"
	cmkClusters     = "/cmk/v2/clusters"
	orgEnvironment  = "/org/v2/environments/{id}"
	orgEnvironments = "/org/v2/environments"
)

var cmkHandlers = map[string]func(*V2Router, *testing.T) func(http.ResponseWriter, *http.Request){
	cmkCluster:  (*V2Router).HandleCmkCluster,
	cmkClusters: (*V2Router).HandleCmkClusters,
}

var orgHandlers = map[string]func(*V2Router, *testing.T) func(http.ResponseWriter, *http.Request){
	orgEnvironment:  (*V2Router).HandleOrgEnvironment,
	orgEnvironments: (*V2Router).HandleOrgEnvironments,
}

type V2Router struct {
	*mux.Router
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
	for route, handler := range cmkHandlers {
		c.HandleFunc(route, handler(c, t))
	}

	for route, handler := range orgHandlers {
		c.HandleFunc(route, handler(c, t))
	}
}
