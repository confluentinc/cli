package test_server

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
	"github.com/gogo/protobuf/proto"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for POST "/cmk/v2/clusters"
func handleCmkKafkaClusterCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		req := new(cmkv2.CmkV2Cluster)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		cluster := &cmkv2.CmkV2Cluster{
			Spec: &cmkv2.CmkV2ClusterSpec{
				DisplayName:            cmkv2.PtrString(*req.Spec.DisplayName),
				Cloud:                  cmkv2.PtrString(*req.Spec.Cloud),
				Region:                 cmkv2.PtrString(*req.Spec.Region),
				Config:                 new(cmkv2.CmkV2ClusterSpecConfigOneOf),
				KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
				HttpEndpoint:           cmkv2.PtrString("https://pkc-endpoint"),
				Availability:           req.Spec.Availability,
			},
			Id:     cmkv2.PtrString("lkc-def963"),
			Status: &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONING"},
		}

		if req.Spec.Config.CmkV2Dedicated != nil {
			cluster.Spec.Config.CmkV2Dedicated = &cmkv2.CmkV2Dedicated{
				Kind: "Dedicated",
				Cku:  req.Spec.Config.CmkV2Dedicated.Cku,
			}
			if *req.Spec.DisplayName == "gcp-byok-test" {
				cluster.Spec.Config.CmkV2Dedicated.EncryptionKey = cmkv2.PtrString("xyz")
			}
			cluster.Status.Cku = cmkv2.PtrInt32(1)
		} else {
			cluster.Spec.Config.CmkV2Basic = &cmkv2.CmkV2Basic{Kind: "Basic"}
		}

		err = json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for "/cmk/v2/clusters"
func handleCmkClusters(t *testing.T) http.HandlerFunc {
	write := func(w http.ResponseWriter, resp proto.Message) {
		type errorer interface {
			GetError() *corev1.Error
		}

		if r, ok := resp.(errorer); ok {
			w.WriteHeader(int(r.GetError().Code))
		}

		b, err := utilv1.MarshalJSONToBytes(resp)
		require.NoError(t, err)

		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Header.Get("Authorization") {
		case "Bearer expired":
			write(w, &schedv1.GetKafkaClustersReply{Error: &corev1.Error{Message: "token is expired", Code: http.StatusUnauthorized}})
		case "Bearer malformed":
			write(w, &schedv1.GetKafkaClustersReply{Error: &corev1.Error{Message: "malformed token", Code: http.StatusBadRequest}})
		case "Bearer invalid":
			// TODO: The response for an invalid token should be 4xx, not 500 (e.g., if you take a working token from devel and try in stag)
			write(w, &schedv1.GetKafkaClustersReply{Error: &corev1.Error{Message: "Token parsing error: crypto/rsa: verification error", Code: http.StatusInternalServerError}})
		}
		if r.Method == http.MethodPost {
			handleCmkKafkaClusterCreate(t)(w, r)
		} else if r.Method == http.MethodGet {
			cluster := cmkv2.CmkV2Cluster{
				Id: cmkv2.PtrString("lkc-123"),
				Spec: &cmkv2.CmkV2ClusterSpec{
					DisplayName: cmkv2.PtrString("abc"),
					Cloud:       cmkv2.PtrString("gcp"),
					Region:      cmkv2.PtrString("us-central1"),
					Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
						CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
					},
					Availability: cmkv2.PtrString("SINGLE_ZONE"),
				},
				Status: &cmkv2.CmkV2ClusterStatus{
					Phase: "PROVISIONING",
				},
			}
			clusterMultizone := cmkv2.CmkV2Cluster{
				Id: cmkv2.PtrString("lkc-456"),
				Spec: &cmkv2.CmkV2ClusterSpec{
					DisplayName: cmkv2.PtrString("def"),
					Cloud:       cmkv2.PtrString("gcp"),
					Region:      cmkv2.PtrString("us-central1"),
					Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
						CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
					},
					Availability: cmkv2.PtrString("MULTI_ZONE"),
				},
				Status: &cmkv2.CmkV2ClusterStatus{
					Phase: "PROVISIONING",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			clusterList := &cmkv2.CmkV2ClusterList{Data: []cmkv2.CmkV2Cluster{cluster, clusterMultizone}}
			err := json.NewEncoder(w).Encode(clusterList)
			require.NoError(t, err)
		}
	}
}

// Handler for "/cmk/v2/clusters/{id}"
func handleCmkCluster(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		switch clusterId {
		case "lkc-describe":
			handleCmkKafkaClusterDescribe(t)(w, r)
		case "lkc-describe-dedicated":
			handleCmkKafkaClusterDescribeDedicated(t)(w, r)
		case "lkc-describe-dedicated-pending":
			handleCmkKafkaClusterDescribeDedicatedPending(t)(w, r)
		case "lkc-describe-dedicated-with-encryption":
			handleCmkKafkaClusterDescribeDedicatedWithEncryption(t)(w, r)
		case "lkc-update":
			handleCmkKafkaClusterUpdateRequest(t)(w, r)
		case "lkc-update-dedicated-expand":
			handleCmkKafkaDedicatedClusterExpansion(t)(w, r)
		case "lkc-update-dedicated-shrink":
			handleCmkKafkaDedicatedClusterShrink(t)(w, r)
		case "lkc-unknown":
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		case "lkc-describe-infinite":
			handleCmkKafkaClusterDescribeInfinite(t)(w, r)
		default:
			handleCmkKafkaClusterGetListDeleteDescribe(t)(w, r)
		}
	}
}

// Handler for GET "/cmk/v2/clusters/{id}"
func handleCmkKafkaClusterDescribe(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkBasicDescribeCluster(id, "kafka-cluster")
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

func handleCmkKafkaClusterDescribeDedicated(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 1)
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-dedicated-pending"
func handleCmkKafkaClusterDescribeDedicatedPending(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 2)
		cluster.Status.Cku = cmkv2.PtrInt32(1)
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-dedicated-with-encryption"
func handleCmkKafkaClusterDescribeDedicatedWithEncryption(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 1)
		cluster.Spec.Config.CmkV2Dedicated.EncryptionKey = cmkv2.PtrString("abc123")
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-infinite
func handleCmkKafkaClusterDescribeInfinite(t *testing.T) http.HandlerFunc {
	return handleCmkKafkaClusterDescribeDedicated(t) // dedicated cluster has infinite storage
}

// Default handler for get, list, delete, describe "/cmk/v2/clusters/{id}"
func handleCmkKafkaClusterGetListDeleteDescribe(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		id := vars["id"]
		if r.Method == http.MethodDelete {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// this is in the body of delete requests
		require.NotEmpty(t, r.URL.Query().Get("environment"))
		cluster := getCmkBasicDescribeCluster(id, "kafka-cluster")
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update"
func handleCmkKafkaClusterUpdateRequest(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// var out []byte
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkBasicDescribeCluster(id, "lkc-update")
			cluster.Status = &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONED"}
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			if req.Spec.Config != nil && req.Spec.Config.CmkV2Dedicated.Cku > 0 {
			} else { //update name
				cluster := getCmkBasicDescribeCluster(*req.Id, *req.Spec.DisplayName)
				err := json.NewEncoder(w).Encode(cluster)
				require.NoError(t, err)
			}
			require.NoError(t, err)
		}
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update-dedicated-expand"
func handleCmkKafkaDedicatedClusterExpansion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-expand", 1)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			cluster := getCmkDedicatedDescribeCluster(*req.Id, *req.Spec.DisplayName, req.Spec.Config.CmkV2Dedicated.Cku)
			cluster.Status.Cku = cmkv2.PtrInt32(1)
			err = json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update-dedicated-shrink"
func handleCmkKafkaDedicatedClusterShrink(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-shrink", 2)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			cluster := getCmkDedicatedDescribeCluster(*req.Id, *req.Spec.DisplayName, req.Spec.Config.CmkV2Dedicated.Cku)
			cluster.Status.Cku = cmkv2.PtrInt32(2)
			err = json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
	}
}
