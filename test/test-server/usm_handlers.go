package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	usmv1 "github.com/confluentinc/ccloud-sdk-go-v2/usm/v1"
)

// Handler for: "/usm/v1/kafka-clusters"
func handleUsmKafkaClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environmentId := r.URL.Query().Get("environment")
		switch r.Method {
		case http.MethodPost:
			handleUsmKafkaClustersCreate(t)(w, r)
		case http.MethodGet:
			handleUsmKafkaClustersList(t, environmentId)(w, r)
		}
	}
}

// Handler for: "/usm/v1/kafka-clusters/{id}"
func handleUsmKafkaCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		environmentId := r.URL.Query().Get("environment")
		switch r.Method {
		case http.MethodDelete:
			handleUsmKafkaClusterDelete(t, id)(w, r)
		case http.MethodGet:
			handleUsmKafkaClusterDescribe(t, id, environmentId)(w, r)
		}
	}
}

// Handler for: "/usm/v1/connect-clusters"
func handleUsmConnectClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		environmentId := r.URL.Query().Get("environment")
		switch r.Method {
		case http.MethodPost:
			handleUsmConnectClustersCreate(t)(w, r)
		case http.MethodGet:
			handleUsmConnectClustersList(t, environmentId)(w, r)
		}
	}
}

// Handler for: "/usm/v1/connect-clusters/{id}"
func handleUsmConnectCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		environmentId := r.URL.Query().Get("environment")
		switch r.Method {
		case http.MethodDelete:
			handleUsmConnectClusterDelete(t, id)(w, r)
		case http.MethodGet:
			handleUsmConnectClusterDescribe(t, id, environmentId)(w, r)
		}
	}
}

func handleUsmKafkaClustersCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usmKafkaCluster := &usmv1.UsmV1KafkaCluster{}
		err := json.NewDecoder(r.Body).Decode(usmKafkaCluster)
		require.NoError(t, err)

		usmKafkaCluster.SetId("usmkc-abc123")

		err = json.NewEncoder(w).Encode(usmKafkaCluster)
		require.NoError(t, err)
	}
}

func handleUsmKafkaClustersList(t *testing.T, environmentId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usmKafkaClusters := &usmv1.UsmV1KafkaClusterList{
			Data: []usmv1.UsmV1KafkaCluster{
				{
					Id:                              usmv1.PtrString("usmkc-abc123"),
					DisplayName:                     usmv1.PtrString("kafka-cluster-1"),
					ConfluentPlatformKafkaClusterId: usmv1.PtrString("4k0R9d1GTS5tI9f4Y2xZ0Q"),
					Cloud:                           usmv1.PtrString("AWS"),
					Region:                          usmv1.PtrString("us-west-2"),
					Environment:                     &usmv1.EnvScopedObjectReference{Id: environmentId},
				},
				{
					Id:                              usmv1.PtrString("usmkc-def456"),
					DisplayName:                     usmv1.PtrString("kafka-cluster-2"),
					ConfluentPlatformKafkaClusterId: usmv1.PtrString("5k0R9d1GTS5tI9f4Y2xZ0Q"),
					Cloud:                           usmv1.PtrString("Azure"),
					Region:                          usmv1.PtrString("eastus"),
					Environment:                     &usmv1.EnvScopedObjectReference{Id: environmentId},
				},
			},
		}
		err := json.NewEncoder(w).Encode(usmKafkaClusters)
		require.NoError(t, err)
	}
}

func handleUsmKafkaClusterDelete(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id == "usmkc-invalid" {
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "the kafka cluster usmkc-invalid was not found")
			require.NoError(t, err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUsmKafkaClusterDescribe(t *testing.T, id, environmentId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id == "usmkc-invalid" {
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "The kafka cluster usmkc-invalid was not found.")
			require.NoError(t, err)
		} else {
			usmKafkaCluster := &usmv1.UsmV1KafkaCluster{
				Id:                              usmv1.PtrString(id),
				DisplayName:                     usmv1.PtrString("kafka-cluster-1"),
				ConfluentPlatformKafkaClusterId: usmv1.PtrString("4k0R9d1GTS5tI9f4Y2xZ0Q"),
				Cloud:                           usmv1.PtrString("AWS"),
				Region:                          usmv1.PtrString("us-west-2"),
				Environment:                     &usmv1.EnvScopedObjectReference{Id: environmentId},
			}
			err := json.NewEncoder(w).Encode(usmKafkaCluster)
			require.NoError(t, err)
		}
	}
}

func handleUsmConnectClustersCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usmConnectCluster := &usmv1.UsmV1ConnectCluster{}
		err := json.NewDecoder(r.Body).Decode(usmConnectCluster)
		require.NoError(t, err)

		usmConnectCluster.SetId("usmcc-abc123")
		if usmConnectCluster.GetCloud() == "" {
			usmConnectCluster.SetCloud("AWS")
			usmConnectCluster.SetRegion("us-west-2")
		}

		err = json.NewEncoder(w).Encode(usmConnectCluster)
		require.NoError(t, err)
	}
}

func handleUsmConnectClustersList(t *testing.T, environmentId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		usmConnectClusters := &usmv1.UsmV1ConnectClusterList{
			Data: []usmv1.UsmV1ConnectCluster{
				{
					Id:                                usmv1.PtrString("usmcc-abc123"),
					ConfluentPlatformConnectClusterId: usmv1.PtrString("connect-group-xyz123"),
					KafkaClusterId:                    usmv1.PtrString("4k0R9d1GTS5tI9f4Y2xZ0Q"),
					Cloud:                             usmv1.PtrString("AWS"),
					Region:                            usmv1.PtrString("us-west-2"),
					Environment:                       &usmv1.EnvScopedObjectReference{Id: environmentId},
				},
				{
					Id:                                usmv1.PtrString("usmcc-def456"),
					ConfluentPlatformConnectClusterId: usmv1.PtrString("connect-group-xyz456"),
					KafkaClusterId:                    usmv1.PtrString("5k0R9d1GTS5tI9f4Y2xZ0Q"),
					Cloud:                             usmv1.PtrString("Azure"),
					Region:                            usmv1.PtrString("eastus"),
					Environment:                       &usmv1.EnvScopedObjectReference{Id: environmentId},
				},
			},
		}
		err := json.NewEncoder(w).Encode(usmConnectClusters)
		require.NoError(t, err)
	}
}

func handleUsmConnectClusterDelete(t *testing.T, id string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id == "usmcc-invalid" {
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "the connect cluster usmcc-invalid was not found")
			require.NoError(t, err)
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func handleUsmConnectClusterDescribe(t *testing.T, id, environmentId string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if id == "usmcc-invalid" {
			w.WriteHeader(http.StatusNotFound)
			err := writeErrorJson(w, "the connect cluster usmcc-invalid was not found")
			require.NoError(t, err)
		} else {
			usmConnectCluster := &usmv1.UsmV1ConnectCluster{
				Id:                                usmv1.PtrString(id),
				ConfluentPlatformConnectClusterId: usmv1.PtrString("connect-group-xyz123"),
				KafkaClusterId:                    usmv1.PtrString("4k0R9d1GTS5tI9f4Y2xZ0Q"),
				Cloud:                             usmv1.PtrString("AWS"),
				Region:                            usmv1.PtrString("us-west-2"),
				Environment:                       &usmv1.EnvScopedObjectReference{Id: environmentId},
			}
			err := json.NewEncoder(w).Encode(usmConnectCluster)
			require.NoError(t, err)
		}
	}
}
