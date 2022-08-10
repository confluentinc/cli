package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	v1 "github.com/confluentinc/ccloud-sdk-go-v2-internal/kafka-quotas/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func handleKafkaClientQuota(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			id := "cq-1234"
			egress := "5000"
			ingress := "2000"
			name := "quotaName"
			description := "quota description"
			resp := v1.KafkaQuotasV1ClientQuota{
				Id:          &id,
				DisplayName: &name,
				Description: &description,
				Throughput: &v1.KafkaQuotasV1Throughput{
					IngressByteRate: &ingress,
					EgressByteRate:  &egress,
				},
				Cluster:     &v1.ObjectReference{Id: "lkc-1234"},
				Principals:  &[]v1.ObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
				Environment: &v1.ObjectReference{Id: "env-1234"},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPatch:
			req := &v1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
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
			resp := v1.KafkaQuotasV1ClientQuotaList{
				Data: []v1.KafkaQuotasV1ClientQuota{
					{
						Id:          &id,
						DisplayName: &name,
						Description: &description,
						Throughput: &v1.KafkaQuotasV1Throughput{
							IngressByteRate: &ingress,
							EgressByteRate:  &egress,
						},
						Cluster:     &v1.ObjectReference{Id: "lkc-1234"},
						Principals:  &[]v1.ObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
						Environment: &v1.ObjectReference{Id: "env-1234"},
					},
					{
						Id:          &id2,
						DisplayName: &name2,
						Description: &description,
						Throughput: &v1.KafkaQuotasV1Throughput{
							IngressByteRate: &ingress,
							EgressByteRate:  &egress,
						},
						Cluster:     &v1.ObjectReference{Id: "lkc-1234"},
						Principals:  &[]v1.ObjectReference{{Id: "sa-4321"}},
						Environment: &v1.ObjectReference{Id: "env-1234"},
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPost:
			id := "cq-1234"
			req := &v1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			resp := v1.KafkaQuotasV1ClientQuota{
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
