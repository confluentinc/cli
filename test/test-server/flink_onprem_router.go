package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

var flinkRoutes = []route{
	{"/cmf/api/v1/environments/{environment}/applications", handleCmfApplications},
	{"/cmf/api/v1/environments", handleCmfEnvironments},
}

func NewFlinkOnPremRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}
	return router
}
