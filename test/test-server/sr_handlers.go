package test_server

import (
	"encoding/json"
	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"net/http"
	"strconv"
	"testing"
)

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
// Handler for: "/subjects/{subject}/versions"
func (s *SRRouter) HandleSRSubjectVersions(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case "POST":
			var req srsdk.RegisterSchemaRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.RegisterSchemaRequest{Id: 1})
			require.NoError(t, err)
		case "GET":
			var versions []int32
			if mux.Vars(r)["subject"] == "testSubject" {
				versions = []int32{1, 2, 3}
			}
			err := json.NewEncoder(w).Encode(versions)
			require.NoError(t, err)
		}

	}
}
// Handler for: "/subjects/{subject}"
func (s *SRRouter) HandleSRSubject(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode([]int32{int32(1), int32(2)})
		require.NoError(t, err)
	}
}
// Handler for: "/subjects/{subject}/versions/{version}"
func (s *SRRouter) HandleSRSubjectVersion(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		switch r.Method {
		case "GET":
			versionStr := vars["version"]
			version64, err := strconv.ParseInt(versionStr, 10, 32)
			require.NoError(t, err)
			subject := vars["subject"]
			err = json.NewEncoder(w).Encode(srsdk.Schema{
				Subject:    subject,
				Version:    int32(version64),
				Id:         1,
				SchemaType: "record",
				References: []srsdk.SchemaReference{{
					Name:    "ref",
					Subject: "payment",
					Version: 1,
				}},
				Schema:     "schema",
			})
			require.NoError(t, err)
		case "DELETE":
			err := json.NewEncoder(w).Encode(int32(1))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/schemas/ids/{id}"
func (s *SRRouter) HandleSRById(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		idStr := vars["id"]
		id64, err := strconv.ParseInt(idStr, 10, 32)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(srsdk.Schema{
			Subject:    subject,
			Version:    1,
			Id:         int32(id64),
			SchemaType: "record",
			References: []srsdk.SchemaReference{{
				Name:    "ref",
				Subject: "payment",
				Version: 1,
			}},
			Schema:     "schema",
		})
		require.NoError(t, err)
	}
}
// Handler for: "/subjects"
func (s *SRRouter) HandleSRSubjects(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		subjects := []string{"subject1", "subject2", "subject3"}
		err := json.NewEncoder(w).Encode(subjects)
		require.NoError(t, err)
	}
}
// Handler for: "/config/{subject}"
func (s *SRRouter) HandleSRSubjectConfig(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req srsdk.ConfigUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
		require.NoError(t, err)
	}
}
// Handler for: "/mode/{subject}"
func (s *SRRouter) HandleSRSubjectMode(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		var req srsdk.ModeUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(srsdk.ModeUpdateRequest{Mode: req.Mode})
		require.NoError(t, err)
	}
}
