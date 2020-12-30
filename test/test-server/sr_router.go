package test_server

import (
	"encoding/json"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/stretchr/testify/require"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
)

// schema registry urls
const (
	get = "/"
	updateTopLevelConfig = "/config"
	updateTopLevelMode 	 = "/mode"

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
}
// Handler for: "/"
func (s *SRRouter) HandleSRGet(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]interface{}{})
		require.NoError(t, err)
	}
}
// Handler for: "/config"
func (s *SRRouter) HandleSRUpdateTopLevelConfig(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req srsdk.ConfigUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
		require.NoError(t, err)
	}
}
// Handler for: "/mode"
func (s *SRRouter) HandleSRUpdateTopLevelMode(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req srsdk.ModeUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		w.Header().Set("Content-Type", "application/json")
		err = json.NewEncoder(w).Encode(srsdk.ModeUpdateRequest{Mode: req.Mode})
		require.NoError(t, err)
	}
}
