package testserver

import (
	"encoding/json"
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"
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
			if req.Spec.GetDisplayName() == "gcp-byok-test" {
				cluster.Spec.Config.CmkV2Dedicated.EncryptionKey = cmkv2.PtrString("xyz")
			}
			if req.Spec.GetDisplayName() == "cck-byok-test" {
				cluster.Spec.Byok = req.Spec.Byok
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
	return func(w http.ResponseWriter, r *http.Request) {
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
		switch mux.Vars(r)["id"] {
		case "lkc-create-topic", "lkc-describe", "lkc-describe-topic":
			handleCmkKafkaClusterDescribe(t)(w, r)
		case "lkc-describe-dedicated":
			handleCmkKafkaClusterDescribeDedicated(t)(w, r)
		case "lkc-describe-dedicated-pending":
			handleCmkKafkaClusterDescribeDedicatedPending(t)(w, r)
		case "lkc-describe-dedicated-with-encryption":
			handleCmkKafkaClusterDescribeDedicatedWithEncryption(t)(w, r)
		case "lkc-describe-infinite":
			handleCmkKafkaClusterDescribeInfinite(t)(w, r)
		case "lkc-update":
			handleCmkKafkaClusterUpdateRequest(t)(w, r)
		case "lkc-update-dedicated-expand":
			handleCmkKafkaDedicatedClusterExpansion(t)(w, r)
		case "lkc-update-dedicated-shrink":
			handleCmkKafkaDedicatedClusterShrink(t)(w, r)
		case "lkc-update-dedicated-shrink-multi":
			handleCmkKafkaDedicatedClusterShrinkMulti(t)(w, r)
		case "lkc-unknown":
			handleCmkKafkaUnknown(t)(w, r)
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
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodGet {
			cluster := getCmkBasicDescribeCluster("lkc-update", "lkc-update")
			cluster.Status = &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONED"}
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			req.Id = cmkv2.PtrString("lkc-update")
			if req.Spec.Config != nil && req.Spec.Config.CmkV2Dedicated.Cku > 0 {
			} else { // update name
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
			cluster := getCmkDedicatedDescribeCluster("lkc-update-dedicated-expand", "lkc-update-dedicated-expand", 1)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			req.Id = cmkv2.PtrString("lkc-update-dedicated-expand")
			if req.Spec.GetDisplayName() == "" { // keep the name unchanged when not specified in request
				req.Spec.SetDisplayName("lkc-update-dedicated-expand")
			}
			cluster := getCmkDedicatedDescribeCluster(req.GetId(), req.Spec.GetDisplayName(), req.Spec.Config.CmkV2Dedicated.Cku)
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
			cluster := getCmkDedicatedDescribeCluster("lkc-update-dedicated-shrink", "lkc-update-dedicated-shrink", 2)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			req.Id = cmkv2.PtrString("lkc-update-dedicated-shrink")
			if req.Spec.GetDisplayName() == "" { // keep the name unchanged when not specified in request
				req.Spec.SetDisplayName("lkc-update-dedicated-shrink")
			}
			cluster := getCmkDedicatedDescribeCluster(req.GetId(), req.Spec.GetDisplayName(), req.Spec.Config.CmkV2Dedicated.Cku)
			cluster.Status.Cku = cmkv2.PtrInt32(2)
			err = json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		}
	}
}

// Handler for GET/PATCH "/cmk/v2/clusters/lkc-update-dedicated-shrink-multi"
func handleCmkKafkaDedicatedClusterShrinkMulti(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-shrink-multi", 3)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			_, err := io.WriteString(w, badRequestErrMsg)
			require.NoError(t, err)
		}
	}
}

func handleCmkKafkaUnknown(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		err := writeResourceNotFoundError(w)
		require.NoError(t, err)
	}
}
