package test_server

import (
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

var cmkHandlers = map[string]func(*V2Router, *testing.T) http.HandlerFunc{
	"/cmk/v2/clusters/{id}": (*V2Router).HandleCmkCluster,
	"/cmk/v2/clusters":      (*V2Router).HandleCmkClusters,
}

var orgHandlers = map[string]func(*testing.T) http.HandlerFunc{
	"/org/v2/environments/{id}": HandleOrgEnvironment,
	"/org/v2/environments":      HandleOrgEnvironments,
}

type V2Router struct {
	*mux.Router
}

func NewV2Router(t *testing.T) *V2Router {
	router := &V2Router{
		Router: mux.NewRouter(),
	}
	router.buildV2Handler(t)
	return router
}

func (c *V2Router) buildV2Handler(t *testing.T) {
	for route, handler := range cmkHandlers {
		c.HandleFunc(route, handler(c, t))
	}

	for route, handler := range orgHandlers {
		c.HandleFunc(route, handler(t))
	}
}
