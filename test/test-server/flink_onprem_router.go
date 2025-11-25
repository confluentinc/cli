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
	{"/cmf/api/v1/environments/{environment}/statements/{stmtName}", handleCmfStatement},
	{"/cmf/api/v1/environments/{environment}/statements/{stmtName}/exceptions", handleCmfStatementExceptions},
	{"/cmf/api/v1/environments/{environment}/statements", handleCmfStatements},
	{"/cmf/api/v1/environments/{envName}/applications/{appName}/savepoints", handleCmfSavepoints},
	{"/cmf/api/v1/environments/{envName}/statements/{stmtName}/savepoints", handleCmfSavepoints},
	{"/cmf/api/v1/environments/{envName}/applications/{appName}/savepoints/{savepointName}", handleCmfSavepoint},
	{"/cmf/api/v1/environments/{envName}/statements/{stmtName}/savepoints/{savepointName}", handleCmfSavepoint},
}

func NewFlinkOnPremRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range flinkRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}
	return router
}
