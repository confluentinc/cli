package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

var schemaRegistryRoutes = []route{
	{"/", handleSRGet},
	{"/config", handleSRUpdateTopLevelConfig},
	{"/mode", handleSRUpdateTopLevelMode},
	{"/compatibility/subjects/{subject}/versions/{version}", handleSRCompatibility},
	{"/subjects/{subject}/versions", handleSRSubjectVersions},
	{"/subjects/{subject}", handleSRSubject},
	{"/subjects/{subject}/versions/{version}", handleSRSubjectVersion},
	{"/schemas", handleSRSchemas},
	{"/schemas/ids/{id}", handleSRById},
	{"/subjects", handleSRSubjects},
	{"/exporters", handleSRExporters},
	{"/exporters/{name}", handleSRExporter},
	{"/exporters/{name}/status", handleSRExporterStatus},
	{"/exporters/{name}/config", handleSRExporterConfig},
	{"/exporters/{name}/pause", handleSRExporterPause},
	{"/exporters/{name}/resume", handleSRExporterResume},
	{"/exporters/{name}/reset", handleSRExporterReset},
	{"/config/{subject}", handleSRSubjectConfig},
	{"/mode/{subject}", handleSRSubjectMode},
	{"/asyncapi", handleSRAsyncApi},
}

func NewSRRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range schemaRegistryRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}
