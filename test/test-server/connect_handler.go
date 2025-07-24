package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	camv1 "github.com/confluentinc/ccloud-sdk-go-v2/cam/v1"
	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
)

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
			}

			artifact.Status = &camv1.CamV1ConnectArtifactStatus{
				Phase: "PROCESSING",
			}

			artifactStore[artifact.GetId()] = *artifact

			err := json.NewEncoder(w).Encode(artifact)
			require.NoError(t, err)
		case http.MethodGet:
			var artifacts []camv1.CamV1ConnectArtifact
			for _, artifact := range artifactStore {
				if artifact.GetId() == "cfa-jar123" {
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

			if id == "cfa-jar123" {
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

			if id == "cfa-zip123" {
				w.WriteHeader(http.StatusNoContent)
			}
		}
	}
}

// Handler for: "/cam/v1/presigned-upload-url"
func handleConnectArtifactUploadUrl(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			uploadUrl := camv1.CamV1PresignedUrl{
				Cloud:       camv1.PtrString("AWS"),
				Environment: camv1.PtrString("env-123456"),
				UploadId:    camv1.PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:   camv1.PtrString(fmt.Sprintf("%s/cam/v1/dummy-presigned-url", TestV2CloudUrl.String())),
			}
			err := json.NewEncoder(w).Encode(uploadUrl)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/cam/v1/dummy-presigned-url"
func handleConnectArtifactUploadFile(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
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
			err := json.NewEncoder(w).Encode(map[string]connectv1.ConnectV1ConnectorExpansion{
				"az-connector":   firstConnectorExpansion,
				"az-connector-2": secondConnectorExpansion,
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
