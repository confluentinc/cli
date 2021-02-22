package test_server

import (
	"testing"

	"github.com/gorilla/mux"
)

// schema registry urls
const (
	get                  = "/"
	updateTopLevelConfig = "/config"
	updateTopLevelMode   = "/mode"
	subjectVersions      = "/subjects/{subject}/versions"
	subject              = "/subjects/{subject}"
	subjectVersion       = "/subjects/{subject}/versions/{version}"
	schemaById           = "/schemas/ids/{id}"
	subjects             = "/subjects"
	subjectLevelConfig   = "/config/{subject}"
	modeSubject          = "/mode/{subject}"
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
	s.HandleFunc(subjectVersions, s.HandleSRSubjectVersions(t))
	s.HandleFunc(subject, s.HandleSRSubject(t))
	s.HandleFunc(subjectVersion, s.HandleSRSubjectVersion(t))
	s.HandleFunc(schemaById, s.HandleSRById(t))
	s.HandleFunc(subjects, s.HandleSRSubjects(t))
	s.HandleFunc(subjectLevelConfig, s.HandleSRSubjectConfig(t))
	s.HandleFunc(modeSubject, s.HandleSRSubjectMode(t))
}
