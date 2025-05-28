package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

var flinkRoutes = []route{
	{"/cmf/api/v1/catalogs/kafka", handleCmfCatalogs},
	{"/cmf/api/v1/catalogs/kafka/{catName}", handleCmfCatalog},
	{"/cmf/api/v1/environments/{environment}/applications", handleCmfApplications},
	{"/cmf/api/v1/environments/{environment}/applications/{application}", handleCmfApplication},
	{"/cmf/api/v1/environments", handleCmfEnvironments},
	{"/cmf/api/v1/environments/{environment}", handleCmfEnvironment},
	{"/cmf/api/v1/environments/{environment}/compute-pools", handleCmfComputePools},
	{"/cmf/api/v1/environments/{environment}/compute-pools/{poolName}", handleCmfComputePool},
}

func NewFlinkOnPremRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}
	return router
}
