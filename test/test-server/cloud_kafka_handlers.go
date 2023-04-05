package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	kafkaquotasv1 "github.com/confluentinc/ccloud-sdk-go-v2/kafka-quotas/v1"
)

var kafkaQuotas = []*kafkaquotasv1.KafkaQuotasV1ClientQuota{
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
}

func handleKafkaClientQuota(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		quotaId := mux.Vars(r)["id"]
		if i := getV2Index(kafkaQuotas, quotaId); i != -1 {
			quota := kafkaQuotas[i]
			switch r.Method {
			case http.MethodGet:
				err := json.NewEncoder(w).Encode(quota)
				require.NoError(t, err)
			case http.MethodDelete:
				_, err := io.WriteString(w, "")
				require.NoError(t, err)
			case http.MethodPatch:
				quotaPatch := &kafkaquotasv1.KafkaQuotasV1ClientQuota{ // make a deep copy so changes don't reflect in subsequent tests
					Id:   kafkaquotasv1.PtrString(quota.GetId()),
					Spec: ptr(quota.GetSpec()),
				}
				req := kafkaquotasv1.KafkaQuotasV1ClientQuota{}
				err := json.NewDecoder(r.Body).Decode(&req)
				require.NoError(t, err)
				quotaPatch.Spec.DisplayName = req.Spec.DisplayName
				quotaPatch.Spec.Description = req.Spec.Description
				quotaPatch.Spec.Throughput = req.Spec.Throughput
				quotaPatch.Spec.Principals = req.Spec.Principals
				err = json.NewEncoder(w).Encode(quotaPatch)
				require.NoError(t, err)
			}
		} else {
			// quota not found
			w.WriteHeader(http.StatusForbidden)
		}
	}
}

func handleKafkaClientQuotas(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			quotaList := &kafkaquotasv1.KafkaQuotasV1ClientQuotaList{Data: getV2List(kafkaQuotas)}
			err := json.NewEncoder(w).Encode(quotaList)
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
