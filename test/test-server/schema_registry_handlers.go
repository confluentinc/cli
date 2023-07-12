package testserver

import (
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	srsdk "github.com/confluentinc/schema-registry-sdk-go"
)

// Handler for: "/"
func handleSRGet(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]any{})
		require.NoError(t, err)
	}
}

// Handler for: "/config"
func handleSRUpdateTopLevelConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var req srsdk.ConfigUpdateRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
			require.NoError(t, err)
		case http.MethodGet:
			res := srsdk.Config{
				CompatibilityLevel: "FULL",
				DefaultMetadata: &srsdk.Metadata{
					Properties: map[string]string{
						"owner": "Bob Jones",
						"email": "bob@acme.com",
					},
				},
			}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/mode"
func handleSRUpdateTopLevelMode(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req srsdk.ModeUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(srsdk.ModeUpdateRequest{Mode: req.Mode})
		require.NoError(t, err)
	}
}

// Handler for: "/subjects/{subject}/versions"
func handleSRSubjectVersions(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
func handleSRSubject(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode([]int32{int32(1), int32(2)})
		require.NoError(t, err)
	}
}

// Handler for: "/subjects/{subject}/versions/{version}"
func handleSRSubjectVersion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		switch r.Method {
		case http.MethodGet:
			versionStr := vars["version"]
			if versionStr == "latest" {
				subject := vars["subject"]
				switch subject {
				case "topic2-value":
					err := json.NewEncoder(w).Encode(srsdk.Schema{
						Subject:    subject,
						Version:    1,
						Id:         1,
						SchemaType: "PROTOBUF",
					})
					require.NoError(t, err)
				default:
					err := json.NewEncoder(w).Encode(srsdk.Schema{
						Subject:    subject,
						Version:    1,
						Id:         1,
						SchemaType: "avro",
						Schema:     `{"doc":"Sample schema to help you get started.","fields":[{"doc":"The int type is a 32-bit signed integer.","name":"my_field1","type":"int"},{"doc":"The double type is a double precision(64-bit) IEEE754 floating-point number.","name":"my_field2","type":"double"},{"doc":"The string is a unicode character sequence.","name":"my_field3","type":"string"}],"name":"sampleRecord","namespace":"com.mycorp.mynamespace","type":"AVRO"}`,
					})
					require.NoError(t, err)
				}
			} else {
				version64, err := strconv.ParseInt(versionStr, 10, 32)
				require.NoError(t, err)
				subject := vars["subject"]
				schema := srsdk.Schema{Subject: subject, Version: int32(version64), SchemaType: "AVRO"}
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
					schema.Schema = `{"schema":1}`
					schema.References = []srsdk.SchemaReference{{
						Name:    "ref",
						Subject: "payment",
						Version: 1,
					}}
				}
				err = json.NewEncoder(w).Encode(schema)
				require.NoError(t, err)
			}
		case http.MethodDelete:
			err := json.NewEncoder(w).Encode(int32(1))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/schemas"
func handleSRSchemas(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subjectPrefix := r.URL.Query().Get("subjectPrefix")
		schemas := []srsdk.Schema{
			{
				Subject: "mysubject-1",
				Version: 1,
				Id:      100001,
			},
			{
				Subject: "mysubject-1",
				Version: 2,
				Id:      100002,
			},
		}
		if subjectPrefix == "" {
			schemas = append(schemas, srsdk.Schema{
				Subject: "mysubject-2",
				Version: 1,
				Id:      100003,
			})
		}

		err := json.NewEncoder(w).Encode(schemas)
		require.NoError(t, err)
	}
}

// Handler for: "/schemas/ids/{id}"
func handleSRById(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		idStr := vars["id"]
		id64, err := strconv.ParseInt(idStr, 10, 32)
		require.NoError(t, err)

		schema := srsdk.Schema{Subject: "my-subject", Version: 1, Id: int32(id64)}
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
		case 1005:
			schema.Schema = `{"schema":1}`
			schema.References = []srsdk.SchemaReference{}
			schema.Ruleset = &srsdk.RuleSet{
				DomainRules: []srsdk.Rule{
					{
						Name: "checkSsnLen",
						Kind: "CONDITION",
						Mode: "WRITE",
						Type: "CEL",
						Expr: "size(message.ssn) == 9",
					},
				},
			}
		default:
			schema.Schema = `{"schema":1}`
			schema.References = []srsdk.SchemaReference{{
				Name:    "ref",
				Subject: "payment",
				Version: 1,
			}}
			schema.Ruleset = nil
		}
		err = json.NewEncoder(w).Encode(schema)
		require.NoError(t, err)
	}
}

// Handler for: "/subjects"
func handleSRSubjects(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		subjects := []string{"subject1", "subject2", "subject3", "topic1-value", "topic2-value"}
		err := json.NewEncoder(w).Encode(subjects)
		require.NoError(t, err)
	}
}

// Handler for: "/exporters"
func handleSRExporters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
func handleSRExporter(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
func handleSRExporterStatus(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		status := srsdk.ExporterStatus{
			Name:  name,
			State: "RUNNING",
		}
		err := json.NewEncoder(w).Encode(status)
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/config"
func handleSRExporterConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := json.NewEncoder(w).Encode(map[string]string{"key1": "value1", "key2": "value2"})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/pause"
func handleSRExporterPause(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/resume"
func handleSRExporterResume(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/exporters/{name}/reset"
func handleSRExporterReset(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		name := vars["name"]
		err := json.NewEncoder(w).Encode(srsdk.UpdateExporterResponse{Name: name})
		require.NoError(t, err)
	}
}

// Handler for: "/config/{subject}"
func handleSRSubjectConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var req srsdk.ConfigUpdateRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(srsdk.ConfigUpdateRequest{Compatibility: req.Compatibility})
			require.NoError(t, err)
		case http.MethodGet:
			res := srsdk.Config{
				CompatibilityLevel: "FORWARD",
				CompatibilityGroup: "application.version",
				DefaultRuleSet: &srsdk.RuleSet{
					DomainRules: []srsdk.Rule{
						{
							Name: "checkSsnLen",
							Kind: "CONDITION",
							Mode: "WRITE",
							Type: "CEL",
							Expr: "size(message.ssn) == 9",
						},
					},
				},
				DefaultMetadata: &srsdk.Metadata{
					Properties: map[string]string{
						"owner": "Bob Jones",
						"email": "bob@acme.com",
					},
				},
			}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/mode/{subject}"
func handleSRSubjectMode(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req srsdk.ModeUpdateRequest
		err := json.NewDecoder(r.Body).Decode(&req)
		require.NoError(t, err)
		err = json.NewEncoder(w).Encode(srsdk.ModeUpdateRequest{Mode: req.Mode})
		require.NoError(t, err)
	}
}

// Handler for: "/compatibility"
func handleSRCompatibility(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := srsdk.CompatibilityCheckResponse{IsCompatible: true}
		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}

// Handler for: "/asyncapi"
func handleSRAsyncApi(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			w.WriteHeader(http.StatusOK)
		}
	}
}

// Handler for: "/catalog/v1/types/tagdefs"
func handleSRTagDefs(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			tagDefs := []srsdk.TagDef{
				{Name: "schema_tag"},
				{Name: "topic_tag"},
			}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(
				tagDefs)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/catalog/v1/entity/tags"
func handleSRTags(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			w.WriteHeader(http.StatusOK)
			var req []srsdk.Tag
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			var res []srsdk.TagResponse
			for _, tag := range req {
				res = append(res, srsdk.TagResponse{
					TypeName:   tag.TypeName,
					EntityType: tag.EntityType,
					EntityName: tag.EntityName,
				})
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet:
			res := []srsdk.TagResponse{
				{
					TypeName:   "schema_tag",
					EntityType: "sr_schema",
					EntityName: "lsrc-1234:.:1",
				},
				{
					TypeName:   "topic_tag",
					EntityType: "kafka_topic",
					EntityName: "lsrc-1234:lkc-asyncapi:topic1",
				},
			}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/catalog/v1/entity"
func handleSRUniqueAttributes(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPut:
			var req srsdk.AtlasEntityWithExtInfo
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			if req.Entity.Attributes["description"] != nil {
				w.WriteHeader(http.StatusOK)
			}
		}
	}
}
