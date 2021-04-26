package test_server

import (
	"io"
	"log"
	"net/http"
	"net/url"
	"testing"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// Handler for POST "/api/clusters"
func (c *CloudRouter) HandleKafkaClusterCreate(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &schedv1.CreateKafkaClusterRequest{}
		err := utilv1.UnmarshalJSON(r.Body, req)
		require.NoError(t, err)
		var b []byte
		if req.Config.Deployment.Sku == productv1.Sku_DEDICATED {
			b, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
				Cluster: &schedv1.KafkaCluster{
					Id:              "lkc-def963",
					AccountId:       req.Config.AccountId,
					Name:            req.Config.Name,
					Cku:             req.Config.Cku,
					Deployment:      &schedv1.Deployment{Sku: productv1.Sku_DEDICATED},
					NetworkIngress:  50 * req.Config.Cku,
					NetworkEgress:   150 * req.Config.Cku,
					Storage:         30000 * req.Config.Cku,
					ServiceProvider: req.Config.ServiceProvider,
					Region:          req.Config.Region,
					Endpoint:        "SASL_SSL://kafka-endpoint",
					ApiEndpoint:     c.kafkaApiUrl,
				},
			})
		} else {
			b, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
				Cluster: &schedv1.KafkaCluster{
					Id:              "lkc-def963",
					AccountId:       req.Config.AccountId,
					Name:            req.Config.Name,
					Deployment:      &schedv1.Deployment{Sku: productv1.Sku_BASIC},
					NetworkIngress:  100,
					NetworkEgress:   100,
					Storage:         5000,
					ServiceProvider: req.Config.ServiceProvider,
					Region:          req.Config.Region,
					Endpoint:        "SASL_SSL://kafka-endpoint",
					ApiEndpoint:     c.kafkaApiUrl,
				},
			})
		}
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/api/clusters/{id}"
func (c *CloudRouter) HandleCluster(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		clusterId := vars["id"]
		switch clusterId {
		case "lkc-describe":
			c.HandleKafkaClusterDescribe(t)(w, r)
		case "lkc-topics", "lkc-no-topics", "lkc-create-topic", "lkc-describe-topic", "lkc-delete-topic", "lkc-acls":
			c.HandleKafkaApiOrRestClusters(t)(w, r)
		case "lkc-describe-dedicated":
			c.HandleKafkaClusterDescribeDedicated(t)(w, r)
		case "lkc-describe-dedicated-pending":
			c.HandleKafkaClusterDescribeDedicatedPending(t)(w, r)
		case "lkc-describe-dedicated-with-encryption":
			c.HandleKafkaClusterDescribeDedicatedWithEncryption(t)(w, r)
		case "lkc-update":
			c.HandleKafkaClusterUpdateRequest(t)(w, r)
		case "lkc-update-dedicated":
			c.HandleKafkaDedicatedClusterUpdate(t)(w, r)
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

// Handler for GET "api/clusters/
func (c *CloudRouter) HandleKafkaClusterDescribe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET "api/clusters/
func (c *CloudRouter) HandleKafkaApiOrRestClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.ApiEndpoint = c.kafkaApiUrl
		u, err := url.Parse(c.kafkaRPUrl)
		if err != nil {
			log.Fatal(err)
		}
		cluster.Endpoint = "SASL_SSL://127.0.0.1:" + u.Port()
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET "/api/clusters/lkc-describe-dedicated"
func (c *CloudRouter) HandleKafkaClusterDescribeDedicated(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.Cku = 1
		cluster.Deployment = &schedv1.Deployment{Sku: productv1.Sku_DEDICATED}
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET "/api/clusters/lkc-describe-dedicated-pending"
func (c *CloudRouter) HandleKafkaClusterDescribeDedicatedPending(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.Cku = 1
		cluster.PendingCku = 2
		cluster.Deployment = &schedv1.Deployment{Sku: productv1.Sku_DEDICATED}
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET "/api/clusters/lkc-describe-dedicated-with-encryption"
func (c *CloudRouter) HandleKafkaClusterDescribeDedicatedWithEncryption(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.Cku = 1
		cluster.EncryptionKeyId = "abc123"
		cluster.Deployment = &schedv1.Deployment{Sku: productv1.Sku_DEDICATED}
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET "/api/clusters/lkc-describe-infinite
func (c *CloudRouter) HandleKafkaClusterDescribeInfinite(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.Storage = -1
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Default handler for get, list, delete, describe "api/clusters/{cluster}"
func (c *CloudRouter) HandleKafkaClusterGetListDeleteDescribe(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]
		if r.Method == "DELETE" {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		// this is in the body of delete requests
		require.NotEmpty(t, r.URL.Query().Get("account_id"))
		// Now return the KafkaCluster with updated ApiEndpoint
		cluster := getBaseDescribeCluster(id, "kafka-cluster")
		cluster.ApiEndpoint = c.kafkaApiUrl
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
			Cluster: cluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for GET/PUT "api/clusters/lkc-update"
func (c *CloudRouter) HandleKafkaClusterUpdateRequest(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		// Describe client call
		var out []byte
		if r.Method == "GET" {
			id := r.URL.Query().Get("id")
			cluster := getBaseDescribeCluster(id, "lkc-update")
			cluster.Status = schedv1.ClusterStatus_UP
			var err error
			out, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
				Cluster: cluster,
			})
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == "PUT" {
			req := &schedv1.UpdateKafkaClusterRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			if req.Cluster.Cku > 0 {
				out, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
					Cluster: nil,
					Error: &corev1.Error{
						Message: "cluster expansion is supported for dedicated clusters only",
					},
				})
			} else {
				cluster := getBaseDescribeCluster(req.Cluster.Id, req.Cluster.Name)
				cluster.Status = schedv1.ClusterStatus_UP
				out, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
					Cluster: cluster,
				})
			}
			require.NoError(t, err)
		}
		_, err := io.WriteString(w, string(out))
		require.NoError(t, err)
	}
}

// Handler for GET/PUT "api/clusters/lkc-update-dedicated"
func (c *CloudRouter) HandleKafkaDedicatedClusterUpdate(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		var out []byte
		if r.Method == "GET" {
			id := r.URL.Query().Get("id")
			var err error
			out, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
				Cluster: &schedv1.KafkaCluster{
					Id:              id,
					Name:            "lkc-update-dedicated",
					Cku:             1,
					Deployment:      &schedv1.Deployment{Sku: productv1.Sku_DEDICATED},
					NetworkIngress:  50,
					NetworkEgress:   150,
					Storage:         30000,
					Status:          schedv1.ClusterStatus_EXPANDING,
					ServiceProvider: "aws",
					Region:          "us-west-2",
					Endpoint:        "SASL_SSL://kafka-endpoint",
					ApiEndpoint:     "http://kafka-api-url",
				},
			})
			require.NoError(t, err)
		}
		// Update client call
		if r.Method == "PUT" {
			req := &schedv1.UpdateKafkaClusterRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			out, err = utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClusterReply{
				Cluster: &schedv1.KafkaCluster{
					Id:              req.Cluster.Id,
					Name:            req.Cluster.Name,
					Cku:             1,
					PendingCku:      req.Cluster.Cku,
					Deployment:      &schedv1.Deployment{Sku: productv1.Sku_DEDICATED},
					NetworkIngress:  50 * req.Cluster.Cku,
					NetworkEgress:   150 * req.Cluster.Cku,
					Storage:         30000 * req.Cluster.Cku,
					Status:          schedv1.ClusterStatus_EXPANDING,
					ServiceProvider: "aws",
					Region:          "us-west-2",
					Endpoint:        "SASL_SSL://kafka-endpoint",
					ApiEndpoint:     "http://kafka-api-url",
				},
			})
			require.NoError(t, err)
		}
		_, err := io.WriteString(w, string(out))
		require.NoError(t, err)
	}
}
