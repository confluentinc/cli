package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	ksqlv2 "github.com/confluentinc/ccloud-sdk-go-v2/ksql/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func ptrTo(s string) *string {
	return &s
}

var _false = false
var _true = true

var ksqlCluster1 = ksqlv2.KsqldbcmV2Cluster{
	Id: ptrTo("lksqlc-ksql5"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ptrTo("account ksql"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-qwert",
			Environment: ptrTo("25"),
		},
		Environment: &ksqlv2.ObjectReference{
			Id: "25",
		},
		UseDetailedProcessingLog: &_true,
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ptrTo("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ptrTo("pksqlc-abcde"),
		Storage:      101,
		Phase:        "PROVISIONING",
	},
}

var ksqlCluster2 = ksqlv2.KsqldbcmV2Cluster{
	Id: ptrTo("lksqlc-woooo"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ptrTo("kay cee queue elle"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-zxcvb",
			Environment: ptrTo("25"),
		},
		Environment: &ksqlv2.ObjectReference{
			Id: "25",
		},
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ptrTo("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ptrTo("pksqlc-ghjkl"),
		Storage:      123,
		Phase:        "PROVISIONING",
	},
}
var ksqlClusterForDetailedProcessingLogFalse = ksqlv2.KsqldbcmV2Cluster{
	Id: ptrTo("lksqlc-woooo"),
	Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
		DisplayName: ptrTo("kay cee queue elle"),
		KafkaCluster: &ksqlv2.ObjectReference{
			Id:          "lkc-zxcvb",
			Environment: ptrTo("25"),
		},
		Environment: &ksqlv2.ObjectReference{
			Id: "25",
		},
		UseDetailedProcessingLog: &_false,
	},
	Status: &ksqlv2.KsqldbcmV2ClusterStatus{
		HttpEndpoint: ptrTo("SASL_SSL://ksql-endpoint"),
		TopicPrefix:  ptrTo("pksqlc-ghjkl"),
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

		switch r.Method {
		case http.MethodGet:
			var cluster ksqlv2.KsqldbcmV2Cluster
			switch id {
			case "lksqlc-ksql1":
				cluster = ksqlv2.KsqldbcmV2Cluster{
					Id: ptrTo("lksqlc-ksql1"),
					Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
						DisplayName: ptrTo("account ksql"),
						KafkaCluster: &ksqlv2.ObjectReference{
							Id:          "lkc-12345",
							Environment: ptrTo("25"),
						},
						Environment: &ksqlv2.ObjectReference{
							Id: "25",
						},
					},
					Status: &ksqlv2.KsqldbcmV2ClusterStatus{
						HttpEndpoint: ptrTo("SASL_SSL://ksql-endpoint"),
						TopicPrefix:  ptrTo("pksqlc-abcde"),
						Storage:      101,
						Phase:        "PROVISIONING",
					},
				}
			case "lksqlc-12345":
				cluster = ksqlv2.KsqldbcmV2Cluster{
					Id: ptrTo("lksqlc-12345"),
					Spec: &ksqlv2.KsqldbcmV2ClusterSpec{
						DisplayName: ptrTo("account ksql"),
						KafkaCluster: &ksqlv2.ObjectReference{
							Id:          "lkc-abcde",
							Environment: ptrTo("25"),
						},
						Environment: &ksqlv2.ObjectReference{
							Id: "25",
						},
						CredentialIdentity: ksqlv2.NewObjectReference("sa-12345", "", ""),
					},
					Status: &ksqlv2.KsqldbcmV2ClusterStatus{
						HttpEndpoint: ptrTo("SASL_SSL://ksql-endpoint"),
						TopicPrefix:  ptrTo("pksqlc-zxcvb"),
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
