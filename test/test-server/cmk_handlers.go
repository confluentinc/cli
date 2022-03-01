package test_server

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
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
func (c *V2Router) HandleKafkaClusterCreate(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var req cmkv2.CmkV2Cluster
		requestBody, err := ioutil.ReadAll(r.Body)
		require.NoError(t, err)
		err = json.Unmarshal(requestBody, &req)
		require.NoError(t, err)
		createCluster := &cmkv2.CmkV2Cluster{}
		if req.Spec.Config.CmkV2Dedicated != nil {
			createCluster = &cmkv2.CmkV2Cluster{
				Spec: &cmkv2.CmkV2ClusterSpec{
					DisplayName: cmkv2.PtrString(*req.Spec.DisplayName),
					Cloud:       cmkv2.PtrString(*req.Spec.Cloud),
					Region:      cmkv2.PtrString(*req.Spec.Region),
					Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
						CmkV2Dedicated: &cmkv2.CmkV2Dedicated{Kind: "Dedicated", Cku: req.Spec.Config.CmkV2Dedicated.Cku},
					},
					KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
					HttpEndpoint:           cmkv2.PtrString("https://pkc-endpoint"),
					Availability:           req.Spec.Availability,
				},
				Id: cmkv2.PtrString("lkc-def963"),
				Status: &cmkv2.CmkV2ClusterStatus{
					Phase: "PROVISIONING",
					Cku:   cmkv2.PtrInt32(1),
				},
			}
		} else {
			createCluster = &cmkv2.CmkV2Cluster{
				Spec: &cmkv2.CmkV2ClusterSpec{
					DisplayName: cmkv2.PtrString(*req.Spec.DisplayName),
					Cloud:       cmkv2.PtrString(*req.Spec.Cloud),
					Region:      cmkv2.PtrString(*req.Spec.Region),
					Config: &cmkv2.CmkV2ClusterSpecConfigOneOf{
						CmkV2Basic: &cmkv2.CmkV2Basic{Kind: "Basic"},
					},
					KafkaBootstrapEndpoint: cmkv2.PtrString("SASL_SSL://kafka-endpoint"),
					HttpEndpoint:           cmkv2.PtrString("https://pkc-endpoint"),
					Availability:           req.Spec.Availability,
				},
				Id: cmkv2.PtrString("lkc-def963"),
				Status: &cmkv2.CmkV2ClusterStatus{
					Phase: "PROVISIONING",
				},
			}
		}
		marshalCmkCluster(t, w, r, createCluster)
	}
}

// Handler for "/cmk/v2/clusters"
func (c *V2Router) HandleCmkClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
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
			c.HandleKafkaClusterCreate(t)(w, r)
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
			b, err := json.Marshal(clusterList)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

// Handler for "/cmk/v2/clusters/{id}"
func (c *V2Router) HandleCmkCluster(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		switch clusterId {
		case "lkc-describe":
			c.HandleKafkaClusterDescribe(t)(w, r)
		case "lkc-topics", "lkc-no-topics", "lkc-create-topic", "lkc-describe-topic", "lkc-delete-topic", "lkc-acls", "lkc-create-topic-kafka-api", "lkc-describe-topic-kafka-api", "lkc-delete-topic-kafka-api", "lkc-groups":
			c.HandleKafkaApiOrRestClusters(t)(w, r)
		case "lkc-describe-dedicated":
			c.HandleKafkaClusterDescribeDedicated(t)(w, r)
		case "lkc-describe-dedicated-pending":
			c.HandleKafkaClusterDescribeDedicatedPending(t)(w, r)
		// case "lkc-describe-dedicated-with-encryption":
		// 	c.HandleKafkaClusterDescribeDedicatedWithEncryption(t)(w, r)
		case "lkc-update":
			c.HandleKafkaClusterUpdateRequest(t)(w, r)
		case "lkc-update-dedicated-expand":
			c.HandleKafkaDedicatedClusterExpansion(t)(w, r)
		case "lkc-update-dedicated-shrink":
			c.HandleKafkaDedicatedClusterShrink(t)(w, r)
		case "lkc-unknown":
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		case "lkc-describe-infinite":
			c.HandleKafkaClusterDescribeInfinite(t)(w, r)
		default:
			c.HandleKafkaClusterGetListDeleteDescribe(t)(w, r)
		}
	}
}

// Handler for GET "/cmk/v2/clusters/{id}"
func (c *V2Router) HandleKafkaClusterDescribe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkBasicDescribeCluster(id, "kafka-cluster")
		marshalCmkCluster(t, w, r, cluster)
	}
}

// Handler for GET "/cmk/v2/clusters"
func (c *V2Router) HandleKafkaApiOrRestClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		cluster := getCmkBasicDescribeCluster(clusterId, "kafka-cluster")
		if !strings.Contains(clusterId, "kafka-api") {
			cluster.Spec.HttpEndpoint = cmkv2.PtrString(c.kafkaRPUrl)
		}
		cluster.Spec.KafkaBootstrapEndpoint = cmkv2.PtrString("SASL_SSL://127.0.0.1:")
		marshalCmkCluster(t, w, r, cluster)
	}
}

func (c *V2Router) HandleKafkaClusterDescribeDedicated(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 1)
		marshalCmkCluster(t, w, r, cluster)
	}
}

// Handler for GET "/cmk/v2/clusters/lkc-describe-dedicated-pending"
func (c *V2Router) HandleKafkaClusterDescribeDedicatedPending(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getCmkDedicatedDescribeCluster(id, "kafka-cluster", 2)
		cluster.Status.Cku = cmkv2.PtrInt32(1)
		marshalCmkCluster(t, w, r, cluster)
	}
}

// encryption not in sdk yet
// Handler for GET "/api/clusters/lkc-describe-dedicated-with-encryption"
// func (c *V2Router) HandleKafkaClusterDescribeDedicatedWithEncryption(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
// 	return func(w http.ResponseWriter, r *http.Request) {
// 		vars := mux.Vars(r)
// 		id := vars["id"]
// 		cluster := getBaseDescribeCluster(id, "kafka-cluster")
// 		cluster.Cku = 1
// 		cluster.EncryptionKeyId = "abc123"
// 		cluster.Deployment = &schedv1.Deployment{Sku: productv1.Sku_DEDICATED}
// 		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
// 			Cluster: cluster,
// 		})
// 		require.NoError(t, err)
// 		_, err = io.WriteString(w, string(b))
// 		require.NoError(t, err)
// 	}
// }

// Handler for GET "/cmk/v2/clusters/lkc-describe-infinite
func (c *V2Router) HandleKafkaClusterDescribeInfinite(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return c.HandleKafkaClusterDescribeDedicated(t) // dedicated cluster has infinite storage
}

// Default handler for get, list, delete, describe "/cmk/v2/clusters/{id}"
func (c *V2Router) HandleKafkaClusterGetListDeleteDescribe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
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
		marshalCmkCluster(t, w, r, cluster)
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update"
func (c *V2Router) HandleKafkaClusterUpdateRequest(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Describe client call
		var out []byte
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkBasicDescribeCluster(id, "lkc-update")
			cluster.Status = &cmkv2.CmkV2ClusterStatus{Phase: "PROVISIONED"}
			marshalCmkCluster(t, w, r, cluster)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			requestBody, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			err = json.Unmarshal(requestBody, &req)
			require.NoError(t, err)
			if req.Spec.Config != nil && req.Spec.Config.CmkV2Dedicated.Cku > 0 {
			} else { //update name
				cluster := getCmkBasicDescribeCluster(*req.Id, *req.Spec.DisplayName)
				marshalCmkCluster(t, w, r, cluster)
			}
			require.NoError(t, err)
		}
		_, err := io.WriteString(w, string(out))
		require.NoError(t, err)
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update-dedicated-expand"
func (c *V2Router) HandleKafkaDedicatedClusterExpansion(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var out []byte
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-expand", 1)
			marshalCmkCluster(t, w, r, cluster)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			requestBody, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			err = json.Unmarshal(requestBody, &req)
			require.NoError(t, err)
			cluster := getCmkDedicatedDescribeCluster(*req.Id, *req.Spec.DisplayName, req.Spec.Config.CmkV2Dedicated.Cku)
			cluster.Status.Cku = cmkv2.PtrInt32(1)
			marshalCmkCluster(t, w, r, cluster)
		}
		_, err := io.WriteString(w, string(out))
		require.NoError(t, err)
	}
}

// Handler for GET/PUT "/cmk/v2/clusters/lkc-update-dedicated-shrink"
func (c *V2Router) HandleKafkaDedicatedClusterShrink(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			id := r.URL.Query().Get("id")
			cluster := getCmkDedicatedDescribeCluster(id, "lkc-update-dedicated-shrink", 2)
			marshalCmkCluster(t, w, r, cluster)
		}
		// Update client call
		if r.Method == http.MethodPatch {
			var req cmkv2.CmkV2Cluster
			requestBody, err := ioutil.ReadAll(r.Body)
			require.NoError(t, err)
			err = json.Unmarshal(requestBody, &req)
			require.NoError(t, err)
			cluster := getCmkDedicatedDescribeCluster(*req.Id, *req.Spec.DisplayName, req.Spec.Config.CmkV2Dedicated.Cku)
			cluster.Status.Cku = cmkv2.PtrInt32(2)
			marshalCmkCluster(t, w, r, cluster)
		}
	}
}

func marshalCmkCluster(t *testing.T, w http.ResponseWriter, r *http.Request, cluster *cmkv2.CmkV2Cluster) http.ResponseWriter {
	w.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(cluster)
	require.NoError(t, err)
	_, err = io.WriteString(w, string(b))
	require.NoError(t, err)
	return w
}
