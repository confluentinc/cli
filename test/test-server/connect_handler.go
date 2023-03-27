package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/stretchr/testify/require"

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
func handleConnectorPause(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/connect/v1/environments/{env}/clusters/{clusters}/connectors/{connector}/resume"
func handleConnectorResume(t *testing.T) http.HandlerFunc {
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
						Trace: connectv1.PtrString(""),
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
						Trace: connectv1.PtrString(""),
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
				Name:   *request.Name,
				Config: *request.Config,
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
