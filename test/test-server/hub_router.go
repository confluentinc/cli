package testserver

import (
	"testing"

	"github.com/gorilla/mux"
)

func NewHubRouter(t *testing.T) *mux.Router {
	router := mux.NewRouter()
	router.Use(defaultHeaderMiddleware)

	router.HandleFunc("/api/plugins/{owner}/{id}", handleHubPlugin(t))
	router.HandleFunc("/api/plugins/{owner}/{id}/versions/{version}", handleHubPluginVersion(t))
	router.HandleFunc("/api/plugins/{owner}/{id}/versions/{version}/{archive}", handleHubPluginArchive(t))

	return router
}
