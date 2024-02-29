package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	connectcustompluginv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect-custom-plugin/v1"
	connectv1 "github.com/confluentinc/ccloud-sdk-go-v2/connect/v1"
)

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
					Tasks: &[]connectv1.ConnectV1ConnectorExpansionStatusTasks{{Id: 1, State: "RUNNING"}},
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
					Tasks: &[]connectv1.ConnectV1ConnectorExpansionStatusTasks{{Id: 1, State: "RUNNING"}},
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
func handleCustomPlugin(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			plugin := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
				Id:          PtrString("ccp-123456"),
				DisplayName: PtrString("my-custom-plugin"),
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		}
		if r.Method == http.MethodGet {
			plugin1 := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
				Id:          PtrString("ccp-123456"),
				DisplayName: PtrString("CliPluginTest1"),
			}
			plugin2 := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
				Id:          PtrString("ccp-789012"),
				DisplayName: PtrString("CliPluginTest2"),
			}
			err := json.NewEncoder(w).Encode(connectcustompluginv1.ConnectV1CustomConnectorPluginList{Data: []connectcustompluginv1.ConnectV1CustomConnectorPlugin{plugin1, plugin2}})
			require.NoError(t, err)
		}
	}
}

// Handler for: "/connect/v1/custom-connector-plugins/{id}"
func handleCustomPluginWithId(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			vars := mux.Vars(r)
			id := vars["id"]
			var plugin connectcustompluginv1.ConnectV1CustomConnectorPlugin
			if id == "ccp-123456" {
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:             PtrString("ccp-123456"),
					DisplayName:    PtrString("CliPluginTest"),
					ConnectorType:  PtrString("source"),
					ConnectorClass: PtrString("io.confluent.kafka.connect.test"),
				}
			} else {
				sensitiveProperties := []string{"aws.key", "aws.secret"}
				plugin = connectcustompluginv1.ConnectV1CustomConnectorPlugin{
					Id:                        PtrString("ccp-123456"),
					DisplayName:               PtrString("CliPluginTest"),
					Description:               PtrString("Source datagen plugin"),
					ConnectorType:             PtrString("source"),
					ConnectorClass:            PtrString("io.confluent.kafka.connect.test"),
					SensitiveConfigProperties: &sensitiveProperties,
				}
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		}
		if r.Method == http.MethodPatch {
			plugin := connectcustompluginv1.ConnectV1CustomConnectorPlugin{
				Id:          PtrString("ccp-123456"),
				DisplayName: PtrString("CliPluginTestUpdate"),
			}
			err := json.NewEncoder(w).Encode(plugin)
			require.NoError(t, err)
		}
		if r.Method == http.MethodDelete {
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
				ContentFormat: PtrString("ZIP"),
				UploadId:      PtrString("e53bb2e8-8de3-49fa-9fb1-4e3fd9a16b66"),
				UploadUrl:     PtrString(fmt.Sprintf("http://%s/connect/v1/dummy-presigned-url", TestV2CloudUrl.Host)),
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
			err := json.NewEncoder(w).Encode(PtrString("Success"))
			require.NoError(t, err)
		}
	}
}
