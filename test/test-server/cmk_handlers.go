package testserver

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	cmkv2 "github.com/confluentinc/ccloud-sdk-go-v2/cmk/v2"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

// Handler for POST "/cmk/v2/clusters"
func handleCmkKafkaClusterCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(cmkv2.CmkV2Cluster)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		cluster := &cmkv2.CmkV2Cluster{
			Spec: &cmkv2.CmkV2ClusterSpec{
				DisplayName:            cmkv2.PtrString(req.Spec.GetDisplayName()),
				Cloud:                  cmkv2.PtrString(req.Spec.GetCloud()),
				Region:                 cmkv2.PtrString(req.Spec.GetRegion()),
				Config:                 &cmkv2.CmkV2ClusterSpecConfigOneOf{},
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
			if req.Spec.GetDisplayName() == "cck-network-test" {
				cluster.Spec.Network = req.Spec.Network
			}
			cluster.Status.Cku = cmkv2.PtrInt32(1)
		} else if req.Spec.Config.CmkV2Enterprise != nil {
			if req.Spec.GetAvailability() == "SINGLE_ZONE" {
				err := writeError(w, "Durability must be HIGH for an Enterprise cluster")
				require.NoError(t, err)
				return
			}
			cluster.Spec.Config.CmkV2Enterprise = &cmkv2.CmkV2Enterprise{Kind: "Enterprise"}
		} else {
			cluster.Spec.Config.CmkV2Basic = &cmkv2.CmkV2Basic{Kind: "Basic"}
		}

		if req.Spec.GetCloud() == "oops" {
			err := writeError(w, "Service provider must be set to AWS, GCP or AZURE.")
			require.NoError(t, err)
			return
		}

		if req.Spec.GetRegion() == "oops" {
			err := writeError(w, "Unable to schedule given the cloud and/or region in request is invalid or unavailable")
			require.NoError(t, err)
			return
		}

		err = json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

func writeError(w http.ResponseWriter, detail string) error {
	w.WriteHeader(http.StatusUnprocessableEntity)
	body := &errors.ErrorResponseBody{Errors: []errors.ErrorDetail{{Detail: detail}}}
	return json.NewEncoder(w).Encode(body)
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
					DisplayName:  cmkv2.PtrString("abc"),
					Cloud:        cmkv2.PtrString("gcp"),
					Region:       cmkv2.PtrString("us-central1"),
					Config:       &cmkv2.CmkV2ClusterSpecConfigOneOf{CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"}},
					Availability: cmkv2.PtrString("SINGLE_ZONE"),
				},
				Status: &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONING"},
			}
			clusterMultizone := cmkv2.CmkV2Cluster{
				Id: cmkv2.PtrString("lkc-456"),
				Spec: &cmkv2.CmkV2ClusterSpec{
					DisplayName:  cmkv2.PtrString("def"),
					Cloud:        cmkv2.PtrString("gcp"),
					Region:       cmkv2.PtrString("us-central1"),
					Config:       &cmkv2.CmkV2ClusterSpecConfigOneOf{CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"}},
					Availability: cmkv2.PtrString("MULTI_ZONE"),
				},
				Status: &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONING"},
			}
			clusterDedicated := getCmkDedicatedDescribeCluster("lkc-789", "ghi", 1)
			clusterDedicated.Spec.Network = &cmkv2.EnvScopedObjectReference{Id: "n-abcde1"}
			clusterList := &cmkv2.CmkV2ClusterList{Data: []cmkv2.CmkV2Cluster{cluster, clusterMultizone, *clusterDedicated}}
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
		case "lkc-describe-dedicated-provisioning":
			handleCmkKafkaClusterDescribeDedicatedProvisioning(t)(w, r)
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
		case "lkc-unknown-type":
			handleCmkKafkaUnknownType(t)(w, r)
		default:
			handleCmkKafkaClusterGetListDeleteDescribe(t)(w, r)
		}
	}
}

// Handler for GET "/cmk/v2/clusters/{id}"
func handleCmkKafkaClusterDescribe(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkBasicDescribeCluster(id, "kafka-cluster")
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

func handleCmkKafkaClusterDescribeDedicated(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 2)
		cluster.Status.Cku = cmkv2.PtrInt32(1)
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-dedicated-provisioning"
func handleCmkKafkaClusterDescribeDedicatedProvisioning(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 1)
		cluster.Status.Phase = "PROVISIONING"
		cluster.Spec.KafkaBootstrapEndpoint = cmkv2.PtrString("")
		cluster.Spec.HttpEndpoint = cmkv2.PtrString("")
		cluster.Spec.Network = &cmkv2.EnvScopedObjectReference{Id: "n-abcde1"}
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-dedicated-with-encryption"
func handleCmkKafkaClusterDescribeDedicatedWithEncryption(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		switch r.Method {
		case http.MethodGet:
			cluster := getCmkBasicDescribeCluster("lkc-update", "lkc-update")
			cluster.Status = &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONED"}
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		case http.MethodPatch:
			var req cmkv2.CmkV2Cluster
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			if req.Spec.Config == nil || req.Spec.Config.CmkV2Dedicated.Cku == 0 {
				req.Id = cmkv2.PtrString("lkc-update")
				cluster := getCmkBasicDescribeCluster(req.GetId(), req.Spec.GetDisplayName())
				err := json.NewEncoder(w).Encode(cluster)
				require.NoError(t, err)
			}
		}
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update-dedicated-expand"
func handleCmkKafkaDedicatedClusterExpansion(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		switch r.Method {
		case http.MethodGet:
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-shrink-multi", 3)
			err := json.NewEncoder(w).Encode(cluster)
			require.NoError(t, err)
		case http.MethodPatch:
			w.WriteHeader(http.StatusBadRequest)
			err := writeErrorJson(w, "Bad Request")
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

// Handler for GET "/cmk/v2/clusters/lkc-unknown-type"
func handleCmkKafkaUnknownType(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkUnknownDescribeCluster(id, "kafka-cluster")
		err := json.NewEncoder(w).Encode(cluster)
		require.NoError(t, err)
	}
}
