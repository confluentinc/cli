package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"
	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
)

type LoggingLogEntry struct {
	Timestamp string            `json:"timestamp"`
	Level     string            `json:"level"`
	Message   string            `json:"message"`
	TaskId    string            `json:"task_id,omitempty"`
	Id        string            `json:"id,omitempty"`
	Exception *LoggingException `json:"exception,omitempty"`
}

type LoggingException struct {
	Stacktrace string `json:"stacktrace,omitempty"`
}
type LoggingMetadata struct {
	Next string `json:"next,omitempty"`
}
type LoggingSearchResponse struct {
	Data       []LoggingLogEntry `json:"data"`
	Metadata   *LoggingMetadata  `json:"metadata,omitempty"`
	ApiVersion string            `json:"api_version"`
	Kind       string            `json:"kind"`
	CRN        string            `json:"crn"`
}
type LoggingSearchParams struct {
	Level      []string `json:"level,omitempty"`
	SearchText string   `json:"search_text,omitempty"`
}

func handleLogsSearch(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		w.Header().Set("Content-Type", "application/json")

		var req struct {
			CRN    string              `json:"crn"`
			Search LoggingSearchParams `json:"search"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		connectorName := ""
		parts := strings.Split(req.CRN, "/")
		for _, part := range parts {
			if strings.HasPrefix(part, "connector=") {
				connectorName = strings.TrimPrefix(part, "connector=")
				break
			}
		}

		if connectorName == "az-connector" {
			response := LoggingSearchResponse{
				Data: []LoggingLogEntry{
					{
						Timestamp: "2025-06-16T05:44:23.761Z",
						Level:     "INFO",
						Message:   "WorkerSourceTask{id=lcc-123-0} Committing offsets for 130 acknowledged messages",
						TaskId:    "task-0",
						Id:        "lcc-123",
					},
					{
						Timestamp: "2025-06-16T05:43:23.757Z",
						Level:     "INFO",
						Message:   "WorkerSourceTask{id=lcc-123-0} Committing offsets for 128 acknowledged messages",
						TaskId:    "task-0",
						Id:        "lcc-123",
					},
					{
						Timestamp: "2025-06-16T05:44:23.761Z",
						Level:     "ERROR",
						Message:   "WorkerSourceTask{id=lcc-123-0} Committing offsets for 130 acknowledged messages",
						TaskId:    "task-0",
						Id:        "lcc-123",
						Exception: &LoggingException{
							Stacktrace: "exception",
						},
					},
				},
				Metadata: &LoggingMetadata{
					Next: "https://api.logging.devel.cpdev.cloud/logs/v1/search?page_token=next-page-token",
				},
				ApiVersion: "v1",
				Kind:       "LoggingSearchResponse",
				CRN:        "crn",
			}
			filteredResponse := LoggingSearchResponse{
				Data: []LoggingLogEntry{},
				Metadata: &LoggingMetadata{
					Next: "https://api.logging.devel.cpdev.cloud/logs/v1/search?page_token=next-page-token",
				},
				ApiVersion: "v1",
				Kind:       "LoggingSearchResponse",
				CRN:        "crn",
			}
			if req.Search.SearchText != "" {
				for _, log := range response.Data {
					if strings.Contains(log.Message, req.Search.SearchText) {
						filteredResponse.Data = append(filteredResponse.Data, log)
					}
				}
			} else {
				filteredResponse = response
			}
			response = filteredResponse
			filteredResponse = LoggingSearchResponse{
				Data: []LoggingLogEntry{},
				Metadata: &LoggingMetadata{
					Next: "https://api.logging.devel.cpdev.cloud/logs/v1/search?page_token=next-page-token",
				},
				ApiVersion: "v1",
				Kind:       "LoggingSearchResponse",
				CRN:        "crn",
			}
			if len(req.Search.Level) > 0 {
				for _, log := range response.Data {
					if slices.Contains(req.Search.Level, log.Level) {
						filteredResponse.Data = append(filteredResponse.Data, log)
					}
				}
			} else {
				filteredResponse = response
			}
			response = filteredResponse
			err := json.NewEncoder(w).Encode(response)
			require.NoError(t, err)
			return
		} else if connectorName == "az-connector-3" {
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		response := LoggingSearchResponse{
			Data: []LoggingLogEntry{},
			Metadata: &LoggingMetadata{
				Next: "https://api.logging.devel.cpdev.cloud/logs/v1/search?page_token=next-page-token",
			},
			ApiVersion: "v1",
			Kind:       "LoggingSearchResponse",
			CRN:        "crn",
		}
		err := json.NewEncoder(w).Encode(response)
		require.NoError(t, err)
	}
}

var artifactStore = make(map[string]camv1.CamV1ConnectArtifact)

// Handler for: "/api/cam/v1/connect-artifacts"
func handleConnectArtifacts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			artifact := &camv1.CamV1ConnectArtifact{}
			require.NoError(t, json.NewDecoder(r.Body).Decode(artifact))

			artifact.Spec.Cloud = strings.ToUpper(artifact.Spec.GetCloud())

			switch artifact.Spec.GetDisplayName() {
			case "my-connect-artifact-jar":
				artifact.SetId("cfa-jar123")
				artifact.Spec.SetContentFormat("JAR")
			case "my-connect-artifact-zip":
				artifact.SetId("cfa-zip123")
				artifact.Spec.SetContentFormat("ZIP")
			case "my-connect-artifact-azure-jar":
				artifact.SetId("cfa-azure-jar123")
				artifact.Spec.SetContentFormat("JAR")
			case "my-connect-artifact-azure-zip":
				artifact.SetId("cfa-azure-zip123")
				artifact.Spec.SetContentFormat("ZIP")
			}

			artifact.Status = &camv1.CamV1ConnectArtifactStatus{
				Phase: "PROCESSING",
			}

			artifactStore[artifact.GetId()] = *artifact

			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
		case http.MethodGet:
			// Try multiple possible query parameter names - the SDK might use different formats
			cloud := ""
			queryParams := r.URL.Query()
			// Check common variations
			for _, param := range []string{"spec.cloud", "cloud", "spec_cloud"} {
				if val := queryParams.Get(param); val != "" {
					cloud = strings.ToUpper(val)
					break
				}
			}
			// Also check all query params in case the name is different
			if cloud == "" {
				for key, values := range queryParams {
					if len(values) > 0 && (strings.Contains(strings.ToLower(key), "cloud") || strings.Contains(strings.ToLower(key), "spec")) {
						cloud = strings.ToUpper(values[0])
						break
					}
				}
			}
			var artifacts []camv1.CamV1ConnectArtifact
			for _, artifact := range artifactStore {
				// Filter by cloud if specified
				if cloud != "" && strings.ToUpper(artifact.Spec.GetCloud()) != cloud {
					continue
				}
				if artifact.GetId() == "cfa-jar123" || artifact.GetId() == "cfa-azure-jar123" {
					artifact.Status = &camv1.CamV1ConnectArtifactStatus{
						Phase: "READY",
					}
				}
				artifacts = append(artifacts, artifact)
			}

			err := json.NewEncoder(w).Encode(camv1.CamV1ConnectArtifactList{Data: artifacts})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/cam/v1/connect-artifacts/{id}"
func handleConnectArtifactId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vars := mux.Vars(r)
			id := vars["id"]
			artifact, exists := artifactStore[id]
			if !exists {
				w.WriteHeader(http.StatusNotFound)
				err := writeErrorJson(w, "The Connect artifact was not found.")
				require.NoError(t, err)
				return
			}

			if id == "cfa-jar123" || id == "cfa-azure-jar123" {
				artifact.Status = &camv1.CamV1ConnectArtifactStatus{
					Phase: "READY",
				}
			}
			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
		case http.MethodDelete:
			vars := mux.Vars(r)
			id := vars["id"]
			if id == "cfa-invalid" {
				w.WriteHeader(http.StatusNotFound)
				err := writeErrorJson(w, "The Connect Artifact was not found.")
				require.NoError(t, err)
				return
			}

			if id == "cfa-zip123" || id == "cfa-azure-zip123" {
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

// Handler for: "/cam/v1/presigned-upload-url"
func handleConnectArtifactUploadUrl(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			var request camv1.CamV1PresignedUrlRequest
			require.NoError(t, json.NewDecoder(r.Body).Decode(&request))

			cloud := strings.ToUpper(request.GetCloud())
			contentFormat := request.GetContentFormat()
			if contentFormat == "" {
				contentFormat = "JAR"
			}

			uploadUrl := camv1.CamV1PresignedUrl{
				Cloud:         camv1.PtrString(cloud),
				Environment:   camv1.PtrString(request.GetEnvironment()),
				UploadId:      camv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:     camv1.PtrString(fmt.Sprintf("%s/cam/v1/dummy-presigned-url", TestV2CloudUrl.String())),
				ContentFormat: camv1.PtrString(strings.ToUpper(contentFormat)),
			}
			err := json.NewEncoder(w).Encode(uploadUrl)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cam/v1/dummy-presigned-url"
func handleConnectArtifactUploadFile(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost || r.Method == http.MethodPut {
			err := json.NewEncoder(w).Encode(camv1.PtrString("Success"))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}"
func handleConnector(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodDelete {
			err := json.NewEncoder(w).Encode(connectv1.InlineResponse200{})
			require.NoError(t, err)
		} else {
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/config"
func handleConnectorConfig(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request map[string]string
		err := json.NewDecoder(r.Body).Decode(&request)
		require.NoError(t, err)
		connector := &connectv1.ConnectV1Connector{
			Name:   "az-connector",
			Config: request,
		}
		err = json.NewEncoder(w).Encode(connector)
		require.NoError(t, err)
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/pause"
func handleConnectorPause(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/resume"
func handleConnectorResume(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

func handleConnectorOffsets(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectorName := strings.Split(r.URL.Path, "/")[8]
		currTime := time.Unix(1712046213, 123).UTC()
		connectorOffset := connectv1.ConnectV1ConnectorOffsets{
			Name: connectv1.PtrString(connectorName),
			Id:   connectv1.PtrString("lcc-123"),
			Offsets: &[]map[string]any{
				0: {
					"partition": map[string]any{
						"server": "dbzv2",
					},
					"offset": map[string]any{
						"event":          2,
						"file":           "mysql-bin.000600",
						"pos":            2001,
						"row":            1,
						"server_id":      1,
						"transaction_id": nil,
						"ts_sec":         1711788870,
					},
				},
			},
			Metadata: &connectv1.ConnectV1ConnectorOffsetsMetadata{
				ObservedAt: &currTime,
			},
		}

		err := json.NewEncoder(w).Encode(connectorOffset)
		require.NoError(t, err)
	}
}

func handleAlterConnectorOffsets(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request connectv1.ConnectV1AlterOffsetRequest
		err := json.NewDecoder(r.Body).Decode(&request)
		if err != nil {
			return
		}
		connectorOffsetRequestInfo := connectv1.ConnectV1AlterOffsetRequestInfo{
			Id:          "lcc-123",
			Name:        "GcsSink",
			Type:        request.Type,
			RequestedAt: time.Unix(1712046213, 123).UTC(),
			Offsets: &[]map[string]any{
				0: {
					"partition": map[string]any{
						"server": "dbzv2",
					},
					"offset": map[string]any{
						"event":          2,
						"file":           "mysql-bin.000600",
						"pos":            2001,
						"row":            1,
						"server_id":      1,
						"transaction_id": nil,
						"ts_sec":         1711788870,
					},
				},
			},
		}

		err = json.NewEncoder(w).Encode(connectorOffsetRequestInfo)
		require.NoError(t, err)
	}
}

func handleAlterConnectorOffsetsStatus(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		connectorName := strings.Split(r.URL.Path, "/")[8]
		currTime := time.Unix(1712046213, 123).UTC()
		var request connectv1.ConnectV1AlterOffsetRequestInfo
		if connectorName == "az-connector" {
			request = connectv1.ConnectV1AlterOffsetRequestInfo{
				Id:          "lcc-123",
				Name:        connectorName,
				Type:        "PATCH",
				RequestedAt: currTime,
				Offsets: &[]map[string]any{
					0: {
						"partition": map[string]any{
							"server": "dbzv2",
						},
						"offset": map[string]any{
							"event":          2,
							"file":           "mysql-bin.000700",
							"pos":            2003,
							"row":            9,
							"server_id":      0,
							"transaction_id": nil,
							"ts_sec":         1711788870,
						},
					},
				},
			}
		} else {
			request = connectv1.ConnectV1AlterOffsetRequestInfo{
				Id:          "lcc-111",
				Name:        connectorName,
				Type:        "DELETE",
				RequestedAt: currTime,
			}
		}

		connectorOffsetStatus := connectv1.ConnectV1AlterOffsetStatus{
			Request: request,
			Status: connectv1.ConnectV1AlterOffsetStatusStatus{
				Phase:   "APPLIED",
				Message: connectv1.PtrString("Offset Updated"),
			},
			PreviousOffsets: &[]map[string]any{
				0: {
					"partition": map[string]any{
						"server": "dbzv2",
					},
					"offset": map[string]any{
						"event":          2,
						"file":           "mysql-bin.000600",
						"pos":            2001,
						"row":            1,
						"server_id":      1,
						"transaction_id": nil,
						"ts_sec":         1711788870,
					},
				},
			},
			AppliedAt: *connectv1.NewNullableTime(&currTime),
		}

		err := json.NewEncoder(w).Encode(connectorOffsetStatus)
		require.NoError(t, err)
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors"
func handleConnectors(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			firstConnectorExpansion := connectv1.ConnectV1ConnectorExpansion{
				Id: &connectv1.ConnectV1ConnectorExpansionId{Id: connectv1.PtrString("lcc-123")},
				Status: &connectv1.ConnectV1ConnectorExpansionStatus{
					Name: "az-connector",
					Connector: connectv1.ConnectV1ConnectorExpansionStatusConnector{
						State: "RUNNING",
					},
					Tasks: &[]connectv1.InlineResponse2001Tasks{{Id: 1, State: "RUNNING"}},
					Type:  "Sink",
				},
				Info: &connectv1.ConnectV1ConnectorExpansionInfo{
					Config: &map[string]string{},
					Name:   connectv1.PtrString("az-connector"),
				},
			}
			secondConnectorExpansion := connectv1.ConnectV1ConnectorExpansion{
				Id: &connectv1.ConnectV1ConnectorExpansionId{Id: connectv1.PtrString("lcc-111")},
				Status: &connectv1.ConnectV1ConnectorExpansionStatus{
					Name: "az-connector-2",
					Connector: connectv1.ConnectV1ConnectorExpansionStatusConnector{
						State: "RUNNING",
					},
					Tasks: &[]connectv1.InlineResponse2001Tasks{{Id: 1, State: "RUNNING"}},
					Type:  "Sink",
				},
				Info: &connectv1.ConnectV1ConnectorExpansionInfo{
					Config: &map[string]string{},
					Name:   connectv1.PtrString("az-connector-2"),
				},
			}
			thirdConnectorExpansion := connectv1.ConnectV1ConnectorExpansion{
				Id: &connectv1.ConnectV1ConnectorExpansionId{Id: connectv1.PtrString("lcc-112")},
				Status: &connectv1.ConnectV1ConnectorExpansionStatus{
					Name: "az-connector-3",
					Connector: connectv1.ConnectV1ConnectorExpansionStatusConnector{
						State: "RUNNING",
					},
					Tasks: &[]connectv1.InlineResponse2001Tasks{{Id: 1, State: "RUNNING"}},
					Type:  "Sink",
				},
				Info: &connectv1.ConnectV1ConnectorExpansionInfo{
					Config: &map[string]string{},
					Name:   connectv1.PtrString("az-connector-3"),
				},
			}
			err := json.NewEncoder(w).Encode(map[string]connectv1.ConnectV1ConnectorExpansion{
				"az-connector":   firstConnectorExpansion,
				"az-connector-2": secondConnectorExpansion,
				"az-connector-3": thirdConnectorExpansion,
			})
			require.NoError(t, err)
		} else if r.Method == http.MethodPost {
			var request connectv1.InlineObject
			err := json.NewDecoder(r.Body).Decode(&request)
			require.NoError(t, err)
			connector := &connectv1.ConnectV1Connector{
				Name:   request.GetName(),
				Config: request.GetConfig(),
			}
			err = json.NewEncoder(w).Encode(connector)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connector-plugins"
func handlePlugins(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			connectorPlugin1 := connectv1.InlineResponse2002{
				Class: "GcsSink",
				Type:  "Sink",
			}
			connectorPlugin2 := connectv1.InlineResponse2002{
				Class: "AzureBlobSink",
				Type:  "Sink",
			}
			err := json.NewEncoder(w).Encode([]connectv1.InlineResponse2002{connectorPlugin1, connectorPlugin2})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connector-plugins/{plugin}/config/validate"
func handlePluginValidate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		configs := &[]connectv1.InlineResponse2003Configs{
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("kafka.api.key"),
					Errors: &[]string{`"kafka.api.key" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Kafka API Key"),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("kafka.api.secret"),
					Errors: &[]string{`"kafka.api.secret" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Kafka API Secret"),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("topics"),
					Errors: &[]string{`"topics" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Identifies the topic name."),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("data.format"),
					Errors: &[]string{`"data.format" is required, Value "null" doesn't belong to the property's "data.format" enum`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Sets the input value format."),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("gcs.credentials.config"),
					Errors: &[]string{`"gcs.credentials.config" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("GCP service account JSON file."),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("gcs.bucket.name"),
					Errors: &[]string{`"gcs.bucket.name" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("GCS bucket name."),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("time.interval"),
					Errors: &[]string{`"data.format" is required, Value "null" doesn't belong to the property's "time.interval" enum`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Partitioning interval of data."),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{
					Name:   connectv1.PtrString("tasks.max"),
					Errors: &[]string{`"tasks.max" is required`},
				},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Tasks"),
					Required:      connectv1.PtrBool(true),
				},
			},
			{
				Value: &connectv1.InlineResponse2003Value{Name: connectv1.PtrString("flush.size")},
				Definition: &connectv1.InlineResponse2003Definition{
					Documentation: connectv1.PtrString("Commit file size."),
					Required:      connectv1.PtrBool(false),
				},
			},
		}

		err := json.NewEncoder(w).Encode(connectv1.InlineResponse2003{Configs: configs})
		require.NoError(t, err)
	}
}

// Handler for: "/connect/v1/custom-connector-plugins"
func handleCustomConnectorPlugins(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			var decodeRespone connectcustompluginv1.ConnectV1CustomConnectorPlugin
			require.NoError(t, json.NewDecoder(r.Body).Decode(&decodeRespone))
			var plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin
			runtimeLanguage := strings.ToLower(decodeRespone.GetRuntimeLanguage())
			cloud := strings.ToLower(decodeRespone.GetCloud())

			if runtimeLanguage == "java" || runtimeLanguage == "" {
				if cloud == "gcp" {
					plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
						Id:             connectcustompluginv1.PtrString("ccp-123456"),
						DisplayName:    connectcustompluginv1.PtrString("my-custom-plugin-gcp"),
						Cloud:          connectcustompluginv1.PtrString("GCP"),
						ConnectorClass: connectcustompluginv1.PtrString("ver-123456"),
						ContentFormat:  connectcustompluginv1.PtrString("JAR"),
					}
				} else if cloud == "azure" {
					plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
						Id:             connectcustompluginv1.PtrString("ccp-123456"),
						DisplayName:    connectcustompluginv1.PtrString("my-custom-plugin-azure"),
						Cloud:          connectcustompluginv1.PtrString("AZURE"),
						ConnectorClass: connectcustompluginv1.PtrString("ver-123456"),
						ContentFormat:  connectcustompluginv1.PtrString("JAR"),
					}
				} else {
					plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
						Id:             connectcustompluginv1.PtrString("ccp-123456"),
						DisplayName:    connectcustompluginv1.PtrString("my-custom-plugin"),
						Cloud:          connectcustompluginv1.PtrString("AWS"),
						ConnectorClass: connectcustompluginv1.PtrString("ver-123456"),
						ContentFormat:  connectcustompluginv1.PtrString("JAR"),
					}
				}
			} else if runtimeLanguage == "python" {
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:             connectcustompluginv1.PtrString("ccp-789012"),
					DisplayName:    connectcustompluginv1.PtrString("my-custom-python-plugin"),
					Cloud:          connectcustompluginv1.PtrString("AWS"),
					ConnectorClass: connectcustompluginv1.PtrString("ver-789012"),
					ContentFormat:  connectcustompluginv1.PtrString("ZIP"),
				}
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodGet:
			customPluginList := &connectcustompluginv1.ConnectV1CustomConnectorPluginList{
				Data: []connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					{
						Id:          connectcustompluginv1.PtrString("ccp-123456"),
						DisplayName: connectcustompluginv1.PtrString("CliPluginTest1"),
						Cloud:       connectcustompluginv1.PtrString("AWS"),
					},
					{
						Id:          connectcustompluginv1.PtrString("ccp-789012"),
						DisplayName: connectcustompluginv1.PtrString("CliPluginTest2"),
						Cloud:       connectcustompluginv1.PtrString("AWS"),
					},
					{
						Id:             connectcustompluginv1.PtrString("ccp-789013"),
						DisplayName:    connectcustompluginv1.PtrString("CliPluginTest3"),
						ConnectorType:  connectcustompluginv1.PtrString("flink_udf"),
						Cloud:          connectcustompluginv1.PtrString("AWS"),
						ConnectorClass: connectcustompluginv1.PtrString("ver_123456"),
						ContentFormat:  connectcustompluginv1.PtrString("JAR"),
					},
				},
			}
			setPageToken(customPluginList, &customPluginList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(customPluginList)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/custom-connector-plugins/{id}"
func handleCustomConnectorPluginsId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			vars := mux.Vars(r)
			id := vars["id"]
			var plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin
			if id == "ccp-123456" {
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:             connectcustompluginv1.PtrString("ccp-123456"),
					DisplayName:    connectcustompluginv1.PtrString("CliPluginTest"),
					ConnectorType:  connectcustompluginv1.PtrString("source"),
					ConnectorClass: connectcustompluginv1.PtrString("io.confluent.kafka.connect.test"),
					Cloud:          connectcustompluginv1.PtrString("AWS"),
				}
			} else if id == "ccp-789012" {
				sensitiveProperties := []string{"aws.key", "aws.secret"}
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:                        connectcustompluginv1.PtrString("ccp-789012"),
					DisplayName:               connectcustompluginv1.PtrString("CliPluginTest"),
					Description:               connectcustompluginv1.PtrString("Source datagen plugin"),
					ConnectorType:             connectcustompluginv1.PtrString("source"),
					ConnectorClass:            connectcustompluginv1.PtrString("io.confluent.kafka.connect.test"),
					Cloud:                     connectcustompluginv1.PtrString("AWS"),
					SensitiveConfigProperties: &sensitiveProperties,
				}
			} else if id == "ccp-401432" {
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:                        connectcustompluginv1.PtrString("ccp-401432"),
					DisplayName:               connectcustompluginv1.PtrString("CliPluginTest"),
					Description:               connectcustompluginv1.PtrString("Source datagen plugin"),
					ConnectorType:             connectcustompluginv1.PtrString("source"),
					ConnectorClass:            connectcustompluginv1.PtrString("io.confluent.kafka.connect.test"),
					Cloud:                     connectcustompluginv1.PtrString("GCP"),
					SensitiveConfigProperties: &[]string{"gcp.key", "gcp.secret"},
				}
			} else {
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:             connectcustompluginv1.PtrString("ccp-789013"),
					DisplayName:    connectcustompluginv1.PtrString("CliPluginTest"),
					ConnectorType:  connectcustompluginv1.PtrString("flink_udf"),
					ConnectorClass: connectcustompluginv1.PtrString("ver-123456"),
					Cloud:          connectcustompluginv1.PtrString("AWS"),
					ContentFormat:  connectcustompluginv1.PtrString("JAR"),
				}
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodPatch:
			plugin := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
				Id:          connectcustompluginv1.PtrString("ccp-123456"),
				DisplayName: connectcustompluginv1.PtrString("CliPluginTestUpdate"),
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		case http.MethodDelete:
			err := json.NewEncoder(w).Encode(connectcustompluginv1.ConnectV1CustomConnectorPlugin{})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/presigned-upload-url"
func handleCustomPluginUploadUrl(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploadUrl := connectcustompluginv1.ConnectV1PresignedUrl{
				ContentFormat: connectcustompluginv1.PtrString("ZIP"),
				Cloud:         connectcustompluginv1.PtrString("AWS"),
				UploadId:      connectcustompluginv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:     connectcustompluginv1.PtrString(fmt.Sprintf("%s/connect/v1/dummy-presigned-url", TestV2CloudUrl.String())),
			}
			err := json.NewEncoder(w).Encode(uploadUrl)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/dummy-presigned-url"
func handleCustomPluginUploadFile(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			err := json.NewEncoder(w).Encode(connectcustompluginv1.PtrString("Success"))
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/custom-connector-runtimes
func handleListCustomConnectorRuntimes(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			runtimes := []connectcustompluginv1.ConnectV1CustomConnectorRuntime{
				{
					Id:                             connectcustompluginv1.PtrString("ccr-123456"),
					CustomConnectPluginRuntimeName: connectcustompluginv1.PtrString("my-custom-runtime"),
					RuntimeAkVersion:               connectcustompluginv1.PtrString("3.7.0"),
					SupportedJavaVersions:          &[]string{"11", "17"},
					ProductMaturity:                connectcustompluginv1.PtrString("GA"),
					EndOfLifeAt:                    connectcustompluginv1.PtrTime(time.Date(2026, 12, 31, 0, 0, 0, 0, time.UTC)),
					Description:                    connectcustompluginv1.PtrString("This is a custom connector runtime for testing purposes."),
				},
				{
					Id:                             connectcustompluginv1.PtrString("ccr-abcdef"),
					CustomConnectPluginRuntimeName: connectcustompluginv1.PtrString("my-custom-runtime"),
					RuntimeAkVersion:               connectcustompluginv1.PtrString("3.8.0"),
					SupportedJavaVersions:          &[]string{"11", "17"},
					ProductMaturity:                connectcustompluginv1.PtrString("EA"),
					Description:                    connectcustompluginv1.PtrString("This is a custom connector runtime for testing purposes."),
				},
			}
			err := json.NewEncoder(w).Encode(connectcustompluginv1.ConnectV1CustomConnectorRuntimeList{Data: runtimes})
			require.NoError(t, err)
		}
	}
}
