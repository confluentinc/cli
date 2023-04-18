package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
)

var ksqlCluster1 = ksqlv2.KsqldbcmV2Cluster{
	Id: ksqlv2.PtrString("lksqlc-ksql5"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ksqlv2.PtrString("account ksql"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-qwert",
			Environment: ksqlv2.PtrString("25"),
		},
		Environment:              &ksqlv2.ObjectReference{Id: "25"},
		UseDetailedProcessingLog: ksqlv2.PtrBool(true),
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ksqlv2.PtrString("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ksqlv2.PtrString("pksqlc-abcde"),
		Storage:      101,
		Phase:        "PROVISIONING",
	},
}

var ksqlCluster2 = ksqlv2.KsqldbcmV2Cluster{
	Id: ksqlv2.PtrString("lksqlc-woooo"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ksqlv2.PtrString("kay cee queue elle"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-zxcvb",
			Environment: ksqlv2.PtrString("25"),
		},
		Environment: &ksqlv2.ObjectReference{Id: "25"},
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ksqlv2.PtrString("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ksqlv2.PtrString("pksqlc-ghjkl"),
		Storage:      123,
		Phase:        "PROVISIONING",
	},
}

var ksqlClusterForDetailedProcessingLogFalse = ksqlv2.KsqldbcmV2Cluster{
	Id: ksqlv2.PtrString("lksqlc-woooo"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ksqlv2.PtrString("kay cee queue elle"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-zxcvb",
			Environment: ksqlv2.PtrString("25"),
		},
		Environment:              &ksqlv2.ObjectReference{Id: "25"},
		UseDetailedProcessingLog: ksqlv2.PtrBool(false),
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ksqlv2.PtrString("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ksqlv2.PtrString("pksqlc-ghjkl"),
		Storage:      123,
		Phase:        "PROVISIONING",
	},
}

// Handler for "/ksqldbcm/v2/clusters"
func handleKsqlClusters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			req := new(ksqlv2.KsqldbcmV2Cluster)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)

			var cluster ksqlv2.KsqldbcmV2Cluster
			if req.Spec.GetUseDetailedProcessingLog() {
				cluster = ksqlCluster1
			} else {
				cluster = ksqlClusterForDetailedProcessingLogFalse
			}

			err = json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		case http.MethodGet:
			clusterList := ksqlv2.KsqldbcmV2ClusterList{
				Data: []ksqlv2.KsqldbcmV2Cluster{
					ksqlCluster1,
					ksqlCluster2,
				},
			}
			err := json.NewEncoder(w).Encode(clusterList)
			require.NoError(t, err)
		default:
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
	}
}

// Handler for "/ksqldbcm/v2/cluster/{id}"
func handleKsqlCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if id != "lksqlc-ksql1" && id != "lksqlc-12345" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			var cluster ksqlv2.KsqldbcmV2Cluster
			switch id {
			case "lksqlc-ksql1":
				cluster = ksqlv2.KsqldbcmV2Cluster{
					Id: ksqlv2.PtrString("lksqlc-ksql1"),
					Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
						DisplayName: ksqlv2.PtrString("account ksql"),
						KafkaCluster: &ksqlv2.ObjectReference{
							Id:          "lkc-12345",
							Environment: ksqlv2.PtrString("25"),
						},
						Environment: &ksqlv2.ObjectReference{Id: "25"},
					},
					Status: &ksqlv2.KsqldbcmV2ClusterStatus{
						HttpEndpoint: ksqlv2.PtrString("SASL_SSL://ksql-endpoint"),
						TopicPrefix:  ksqlv2.PtrString("pksqlc-abcde"),
						Storage:      101,
						Phase:        "PROVISIONING",
					},
				}
			case "lksqlc-12345":
				cluster = ksqlv2.KsqldbcmV2Cluster{
					Id: ksqlv2.PtrString("lksqlc-12345"),
					Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
						DisplayName: ksqlv2.PtrString("account ksql"),
						KafkaCluster: &ksqlv2.ObjectReference{
							Id:          "lkc-abcde",
							Environment: ksqlv2.PtrString("25"),
						},
						Environment:        &ksqlv2.ObjectReference{Id: "25"},
						CredentialIdentity: ksqlv2.NewObjectReference("sa-12345", "", ""),
					},
					Status: &ksqlv2.KsqldbcmV2ClusterStatus{
						HttpEndpoint: ksqlv2.PtrString("SASL_SSL://ksql-endpoint"),
						TopicPrefix:  ksqlv2.PtrString("pksqlc-zxcvb"),
						Storage:      130,
						Phase:        "PROVISIONING",
					},
				}
			}

			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusAccepted)
		default:
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
	}
}
