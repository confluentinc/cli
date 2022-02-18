package test_server

import (
	"testing"

	"github.com/gorilla/mux"
)

// cmk urls
const (
	clusterList = "/cmk/v2/clusters?environment={environment}"
)

type CmkRouter struct {
	router     *mux.Router
	kafkaRPUrl string
}

func NewCmkRouter(t *testing.T) *CmkRouter {
	router := &CmkRouter{router: mux.NewRouter()}
	router.buildCmkHandler(t)
	return router
}

func (c *CmkRouter) buildCmkHandler(t *testing.T) {
	c.router.HandleFunc(clusterList, c.HandleCmkCluster(t))
}
