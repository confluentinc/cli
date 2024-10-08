package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

var flinkRoutes = []route{
	{"/cmf/api/v1/environments/{environment}/applications", handleCmfApplications},
	{"/cmf/api/v1/environments/{environment}/applications/{application}", handleCmfApplication},
	{"/cmf/api/v1/environments", handleCmfEnvironments},
	{"/cmf/api/v1/environments/{environment}", handleCmfEnvironment},
}

func NewFlinkOnPremRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}
	return router
}
