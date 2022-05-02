package test_server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
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
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			var req srsdk.ConfigUpdateRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
			require.NoError(t, err)
		case http.MethodGet:
			res := srsdk.Config{CompatibilityLevel: "FULL"}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
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
		case http.MethodPost:
			var req srsdk.RegisterSchemaRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.RegisterSchemaRequest{Id: 1})
			require.NoError(t, err)
		case http.MethodGet:
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
		case http.MethodGet:
			versionStr := vars["version"]
			version64, err := strconv.ParseInt(versionStr, 10, 32)
			require.NoError(t, err)
			subject := vars["subject"]
			schema := srsdk.Schema{Subject: subject, Version: int32(version64), SchemaType: "record"}
			switch subject {
			case "lvl0":
				schema.Id = 1001
				schema.Schema = "schema0"
				schema.References = []srsdk.SchemaReference{
					{
						Name:    "ref_lvl1_1",
						Subject: "lvl1-1",
						Version: 1,
					},
					{
						Name:    "ref_lvl1_2",
						Subject: "lvl1-2",
						Version: 1,
					},
				}
			case "lvl1-1":
				schema.Id = 1002
				schema.Schema = "schema11"
				schema.References = []srsdk.SchemaReference{
					{
						Name:    "ref_lvl2",
						Subject: "lvl2",
						Version: 1,
					},
				}
			case "lvl1-2":
				schema.Id = 1003
				schema.Schema = "schema12"
				schema.References = []srsdk.SchemaReference{
					{
						Name:    "ref_lvl2",
						Subject: "lvl2",
						Version: 1,
					},
				}
			case "lvl2":
				schema.Id = 1004
				schema.Schema = "schema2"
				schema.References = []srsdk.SchemaReference{}
			default:
				schema.Id = 10
				schema.Schema = "schema"
				schema.References = []srsdk.SchemaReference{{
					Name:    "ref",
					Subject: "payment",
					Version: 1,
				}}
			}
			err = json.NewEncoder(w).Encode(schema)
			require.NoError(t, err)
		case http.MethodDelete:
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

		schema := srsdk.Schema{Subject: subject, Version: 1, SchemaType: "record", Id: int32(id64)}
		switch id64 {
		case 1001:
			schema.Schema = "schema0"
			schema.References = []srsdk.SchemaReference{
				{
					Name:    "ref_lvl1_1",
					Subject: "lvl1-1",
					Version: 1,
				},
				{
					Name:    "ref_lvl1_2",
					Subject: "lvl1-2",
					Version: 1,
				},
			}
		case 1002:
			schema.Schema = "schema11"
			schema.References = []srsdk.SchemaReference{
				{
					Name:    "ref_lvl2",
					Subject: "lvl2",
					Version: 1,
				},
			}
		case 1003:
			schema.Schema = "schema12"
			schema.References = []srsdk.SchemaReference{
				{
					Name:    "ref_lvl2",
					Subject: "lvl2",
					Version: 1,
				},
			}
		case 1004:
			schema.Schema = "schema2"
			schema.References = []srsdk.SchemaReference{}
		default:
			schema.Schema = "schema"
			schema.References = []srsdk.SchemaReference{{
				Name:    "ref",
				Subject: "payment",
				Version: 1,
			}}
		}
		err = json.NewEncoder(w).Encode(schema)
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

// Handler for: "/exporters"
func (s *SRRouter) HandleSRExporters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			exporters := []string{"exporter1", "exporter2"}
			err := json.NewEncoder(w).Encode(exporters)
			require.NoError(t, err)
		case http.MethodPost:
			var req srsdk.CreateExporterRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.CreateExporterResponse{Name: req.Name})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/exporters/{name}"
func (s *SRRouter) HandleSRExporter(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		switch r.Method {
		case http.MethodGet:
			info := srsdk.ExporterInfo{
				Name:                name,
				Subjects:            []string{"foo", "bar"},
				ContextType:         "CUSTOM",
				Context:             "mycontext",
				SubjectRenameFormat: "my-${subject}",
				Config:              map[string]string{"key1": "value1", "key2": "value2"},
			}
			err := json.NewEncoder(w).Encode(info)
			require.NoError(t, err)
		case http.MethodPut:
			err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/exporters/{name}/status"
func (s *SRRouter) HandleSRExporterStatus(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		status := srsdk.ExporterStatus{
			Name:   name,
			State:  "RUNNING",
			Offset: 0,
			Ts:     0,
			Trace:  "",
		}
		err := json.NewEncoder(w).Encode(status)
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/config"
func (s *SRRouter) HandleSRExporterConfig(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		err := json.NewEncoder(w).Encode(map[string]string{"key1": "value1", "key2": "value2"})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/pause"
func (s *SRRouter) HandleSRExporterPause(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/resume"
func (s *SRRouter) HandleSRExporterResume(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/reset"
func (s *SRRouter) HandleSRExporterReset(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/config/{subject}"
func (s *SRRouter) HandleSRSubjectConfig(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodPut:
			var req srsdk.ConfigUpdateRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
			require.NoError(t, err)
		case http.MethodGet:
			res := srsdk.Config{CompatibilityLevel: "FORWARD"}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
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

// Handler for: "/compatibility"
func (c *SRRouter) HandleSRCompatibility(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		res := srsdk.CompatibilityCheckResponse{IsCompatible: true}
		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}
