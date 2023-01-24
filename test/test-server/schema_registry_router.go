package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

// Schema Registry URLs
const (
	get                  = "/"
	updateTopLevelConfig = "/config"
	updateTopLevelMode   = "/mode"
	compatibility        = "/compatibility/subjects/{subject}/versions/{version}"
	subjectVersions      = "/subjects/{subject}/versions"
	subject              = "/subjects/{subject}"
	subjectVersion       = "/subjects/{subject}/versions/{version}"
	schemas              = "/schemas"
	schemaById           = "/schemas/ids/{id}"
	subjects             = "/subjects"
	exporters            = "/exporters"
	exporter             = "/exporters/{name}"
	exporterStatus       = "/exporters/{name}/status"
	exporterConfig       = "/exporters/{name}/config"
	exporterPause        = "/exporters/{name}/pause"
	exporterResume       = "/exporters/{name}/resume"
	exporterReset        = "/exporters/{name}/reset"
	subjectLevelConfig   = "/config/{subject}"
	modeSubject          = "/mode/{subject}"
	asyncApi             = "/asyncapi"
)

type SRRouter struct {
	*mux.Router
}

func NewSRRouter(t *testing.T) *SRRouter {
	router := NewEmptySRRouter()
	router.buildSRHandler(t)
	return router
}

func NewEmptySRRouter() *SRRouter {
	return &SRRouter{
		mux.NewRouter(),
	}
}

func (s *SRRouter) buildSRHandler(t *testing.T) {
	s.HandleFunc(get, s.HandleSRGet(t))
	s.HandleFunc(updateTopLevelConfig, s.HandleSRUpdateTopLevelConfig(t))
	s.HandleFunc(updateTopLevelMode, s.HandleSRUpdateTopLevelMode(t))
	s.HandleFunc(compatibility, s.HandleSRCompatibility(t))
	s.HandleFunc(subjectVersions, s.HandleSRSubjectVersions(t))
	s.HandleFunc(subject, s.HandleSRSubject(t))
	s.HandleFunc(subjectVersion, s.HandleSRSubjectVersion(t))
	s.HandleFunc(schemas, s.HandleSRSchemas(t))
	s.HandleFunc(schemaById, s.HandleSRById(t))
	s.HandleFunc(subjects, s.HandleSRSubjects(t))
	s.HandleFunc(exporters, s.HandleSRExporters(t))
	s.HandleFunc(exporter, s.HandleSRExporter(t))
	s.HandleFunc(exporterStatus, s.HandleSRExporterStatus(t))
	s.HandleFunc(exporterConfig, s.HandleSRExporterConfig(t))
	s.HandleFunc(exporterPause, s.HandleSRExporterPause(t))
	s.HandleFunc(exporterResume, s.HandleSRExporterResume(t))
	s.HandleFunc(exporterReset, s.HandleSRExporterReset(t))
	s.HandleFunc(subjectLevelConfig, s.HandleSRSubjectConfig(t))
	s.HandleFunc(modeSubject, s.HandleSRSubjectMode(t))
	s.HandleFunc(asyncApi, s.HandleSRAsyncApi(t))
}
