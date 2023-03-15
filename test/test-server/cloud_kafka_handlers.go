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
		switch r.Method {
		case http.MethodGet:
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuota{
				Id: kafkaquotasv1.PtrString("cq-1234"),
				Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
					DisplayName: kafkaquotasv1.PtrString("quotaName"),
					Description: kafkaquotasv1.PtrString("quota description"),
					Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
						IngressByteRate: "2000",
						EgressByteRate:  "5000",
					},
					Cluster:     &kafkaquotasv1.EnvScopedObjectReference{Id: "lkc-1234"},
					Principals:  &[]kafkaquotasv1.GlobalObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
					Environment: &kafkaquotasv1.GlobalObjectReference{Id: "env-1234"},
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPatch:
			req := &kafkaquotasv1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			req.Spec.Cluster = &kafkaquotasv1.EnvScopedObjectReference{Id: "lkc-1234"}
			req.Spec.Environment = &kafkaquotasv1.GlobalObjectReference{Id: "env-1234"}
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
		switch r.Method {
		case http.MethodGet:
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuotaList{
				Data: []kafkaquotasv1.KafkaQuotasV1ClientQuota{
					{
						Id: kafkaquotasv1.PtrString("cq-1234"),
						Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
							DisplayName: kafkaquotasv1.PtrString("quotaName"),
							Description: kafkaquotasv1.PtrString("quota description"),
							Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
								IngressByteRate: "2000",
								EgressByteRate:  "5000",
							},
							Cluster:     &kafkaquotasv1.EnvScopedObjectReference{Id: "lkc-1234"},
							Principals:  &[]kafkaquotasv1.GlobalObjectReference{{Id: "sa-1234"}, {Id: "sa-5678"}},
							Environment: &kafkaquotasv1.GlobalObjectReference{Id: "env-1234"},
						},
					},
					{
						Id: kafkaquotasv1.PtrString("cq-4321"),
						Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
							DisplayName: kafkaquotasv1.PtrString("quota2"),
							Description: kafkaquotasv1.PtrString("quota description"),
							Throughput: &kafkaquotasv1.KafkaQuotasV1Throughput{
								IngressByteRate: "2000",
								EgressByteRate:  "5000",
							},
							Cluster:     &kafkaquotasv1.EnvScopedObjectReference{Id: "lkc-1234"},
							Principals:  &[]kafkaquotasv1.GlobalObjectReference{{Id: "sa-4321"}},
							Environment: &kafkaquotasv1.GlobalObjectReference{Id: "env-1234"},
						},
					},
				},
			}
			err := json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		case http.MethodPost:
			req := &kafkaquotasv1.KafkaQuotasV1ClientQuota{}
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			resp := kafkaquotasv1.KafkaQuotasV1ClientQuota{
				Id: kafkaquotasv1.PtrString("cq-1234"),
				Spec: &kafkaquotasv1.KafkaQuotasV1ClientQuotaSpec{
					DisplayName: req.Spec.DisplayName,
					Description: req.Spec.Description,
					Throughput:  req.Spec.Throughput,
					Cluster:     req.Spec.Cluster,
					Principals:  req.Spec.Principals,
					Environment: req.Spec.Environment,
				},
			}
			err = json.NewEncoder(w).Encode(resp)
			require.NoError(t, err)
		}
	}
}
