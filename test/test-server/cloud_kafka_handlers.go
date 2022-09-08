package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func handleKafkaClientQuota(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuota{
				Id:          kafkaquotasv1.PtrString("cq-1234"),
				DisplayName: kafkaquotasv1.PtrString("quotaName"),
				Description: kafkaquotasv1.PtrString("quota description"),
				Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
					IngressByteRate: kafkaquotasv1.PtrString("2000"),
					EgressByteRate:  kafkaquotasv1.PtrString("5000"),
				},
				Cluster:     &kafkaquotasv1.ObjectReference{Id: "lkc-1234"},
				Principals:  &[]kafkaquotasv1.ObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
				Environment: &kafkaquotasv1.ObjectReference{Id: "env-1234"},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPatch:
			req := &kafkaquotasv1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			req.Cluster = &kafkaquotasv1.ObjectReference{Id: "lkc-1234"}
			req.Environment = &kafkaquotasv1.ObjectReference{Id: "env-1234"}
			err = json.NewEncoder(w).Encode(req)
			require.NoError(t, err)
		case http.MethodDelete:
			idStr := mux.Vars(r)["id"]
			require.Equal(t, "cq-123", idStr)
		}
	}
}

func handleKafkaClientQuotas(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			id := "cq-1234"
			id2 := "cq-4321"
			egress := "5000"
			ingress := "2000"
			name := "quotaName"
			name2 := "quota2"
			description := "quota description"
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuotaList{
				Data: []kafkaquotasv1.KafkaQuotasV1ClientQuota{
					{
						Id:          &id,
						DisplayName: &name,
						Description: &description,
						Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
							IngressByteRate: &ingress,
							EgressByteRate:  &egress,
						},
						Cluster:     &kafkaquotasv1.ObjectReference{Id: "lkc-1234"},
						Principals:  &[]kafkaquotasv1.ObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
						Environment: &kafkaquotasv1.ObjectReference{Id: "env-1234"},
					},
					{
						Id:          &id2,
						DisplayName: &name2,
						Description: &description,
						Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
							IngressByteRate: &ingress,
							EgressByteRate:  &egress,
						},
						Cluster:     &kafkaquotasv1.ObjectReference{Id: "lkc-1234"},
						Principals:  &[]kafkaquotasv1.ObjectReference{{Id: "sa-4321"}},
						Environment: &kafkaquotasv1.ObjectReference{Id: "env-1234"},
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPost:
			id := "cq-1234"
			req := &kafkaquotasv1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuota{
				Id:          &id,
				DisplayName: req.DisplayName,
				Description: req.Description,
				Throughput:  req.Throughput,
				Cluster:     req.Cluster,
				Principals:  req.Principals,
				Environment: req.Environment,
			}
			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}
	}
}
