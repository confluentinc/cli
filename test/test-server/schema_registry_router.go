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
	{"/dek-registry/v1/keks", handleSRKeks},
	{"/dek-registry/v1/keks/{name}", handleSRKek},
	{"/dek-registry/v1/keks/{name}/undelete", handleSRKekUndelete},
	{"/dek-registry/v1/keks/{name}/deks", handleSRDeks},
	{"/dek-registry/v1/keks/{name}/deks/{subject}", handleSRDekSubject},
	{"/dek-registry/v1/keks/{name}/deks/{subject}/versions", handleSRDekVersions},
	{"/dek-registry/v1/keks/{name}/deks/{subject}/versions/{version}", handleSRDekVersion},
	{"/dek-registry/v1/keks/{name}/deks/{subject}/versions/{version}/undelete", handleSRDekUndelete},
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
	{"/catalog/v1/types/tagdefs", handleSRTagDefs},
	{"/catalog/v1/entity/tags", handleSRTags},
	{"/catalog/v1/entity", handleSRUniqueAttributes},
}

func NewSRRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	for _, route := range schemaRegistryRoutes {
		router.HandleFunc(route.path, route.handler(t))
	}

	return router
}
