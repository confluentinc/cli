package test_server

import (
	"encoding/json"
	"fmt"
	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

var (
	environments  = []*orgv1.Account{{Id: "a-595", Name: "default"}, {Id: "not-595", Name: "other"}, {Id: "env-123", Name: "env123"}}
)
// Handler for: "/api/me"
func (c *CloudRouter) HandleMe(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := utilv1.MarshalJSONToBytes(&orgv1.GetUserReply{
			User: &orgv1.User{
				Id:         23,
				Email:      "cody@confluent.io",
				FirstName:  "Cody",
				ResourceId: "u-11aaa",
			},
			Accounts: environments,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
// Handler for: "/api/sessions"
func (c *CloudRouter) HandleLogin(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := require.New(t)
		b, err := ioutil.ReadAll(r.Body)
		req.NoError(err)
		auth := &struct {
			Email    string
			Password string
		}{}
		err = json.Unmarshal(b, auth)
		req.NoError(err)
		switch auth.Email {
		case "incorrect@user.com":
			w.WriteHeader(http.StatusForbidden)
		case "expired@user.com":
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1MzAxMjQ4NTcsImV4cCI6MTUzMDAzODQ1NywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSJ9.Y2ui08GPxxuV9edXUBq-JKr1VPpMSnhjSFySczCby7Y"})
		case "malformed@user.com":
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "malformed"})
		case "invalid@user.com":
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "invalid"})
		default:
			http.SetCookie(w, &http.Cookie{Name: "auth_token", Value: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"})
		}
	}
}
// Handler for: "/api/check_email/{email}"
func (c *CloudRouter) HandleCheckEmail(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := require.New(t)
		vars := mux.Vars(r)
		email := vars["email"]
		reply := &orgv1.GetUserReply{}
		switch email {
		case "cody@confluent.io":
			reply.User = &orgv1.User{
				Email: "cody@confluent.io",
			}
		}
		b, err := utilv1.MarshalJSONToBytes(reply)
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	}
}
// Handler for: "/api/accounts/{id}"
func (c *CloudRouter) HandleEnvironmentRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		for _, env := range environments {
			if env.Id == envId {
				// env found
				if r.Method == "GET" {
					b, err := utilv1.MarshalJSONToBytes(&orgv1.GetAccountReply{Account: env})
					require.NoError(t, err)
					_, err = io.WriteString(w, string(b))
					require.NoError(t, err)
				} else if r.Method == "PUT" {
					req := &orgv1.UpdateAccountRequest{}
					err := utilv1.UnmarshalJSON(r.Body, req)
					require.NoError(t, err)
					env.Name = req.Account.Name
					b, err := utilv1.MarshalJSONToBytes(&orgv1.UpdateAccountReply{Account: env})
					require.NoError(t, err)
					_, err = io.WriteString(w, string(b))
					require.NoError(t, err)
				} else if r.Method == "DELETE" {
					b, err := utilv1.MarshalJSONToBytes(&orgv1.DeleteAccountReply{})
					require.NoError(t, err)
					_, err = io.WriteString(w, string(b))
					require.NoError(t, err)
					_, err = io.WriteString(w, string(b))
					require.NoError(t, err)
				}
				return
			}
		}
		// env not found
		w.WriteHeader(http.StatusNotFound)
	}
}
// Handler for: "/api/accounts"
func (c *CloudRouter) HandleEnvironmentsRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			b, err := utilv1.MarshalJSONToBytes(&orgv1.ListAccountsReply{Accounts: environments})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == "POST" {
			req := &orgv1.CreateAccountRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			account := &orgv1.Account{
				Id:             "a-5555",
				Name:           req.Account.Name,
				OrganizationId: 0,
			}
			b, err := utilv1.MarshalJSONToBytes(&orgv1.CreateAccountReply{
				Account: account,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

func (c *CloudRouter) HandlePaymentInfo(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := orgv1.GetPaymentInfoReply{
			Card: &orgv1.Card{
				Cardholder: "Miles Todzo",
				Brand:      "Visa",
				Last4:      "4242",
				ExpMonth:   "01",
				ExpYear:    "99",
			},
			Organization: &orgv1.Organization{
				Id: 0,
			},
			Error: nil,
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

const (
	exampleAvailability = "low"
	exampleCloud        = "aws"
	exampleClusterType  = "basic"
	exampleMetric       = "ConnectNumRecords"
	exampleNetworkType  = "internet"
	examplePrice        = 1
	exampleRegion       = "us-east-1"
	exampleUnit         = "GB"
)
// Handler for "/api/organizations/
func (c *CloudRouter) HandlePriceTable(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prices := map[string]float64{
			strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
		}

		res := &orgv1.GetPriceTableReply{
			PriceTable: &orgv1.PriceTable{
				PriceTable: map[string]*orgv1.UnitPrices{
					exampleMetric: {Unit: exampleUnit, Prices: prices},
				},
			},
		}

		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

var (
	serviceAccountID  = int32(12345)
	deactivatedUserID = int32(6666)
)

// Handler for: "/api/service_accounts"
func (c *CloudRouter) HandleServiceAccountRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			serviceAccount := &orgv1.User{
				Id:                 serviceAccountID,
				ServiceName:        "service_account",
				ServiceDescription: "at your service.",
			}
			listReply, err := utilv1.MarshalJSONToBytes(&orgv1.GetServiceAccountsReply{
				Users: []*orgv1.User{serviceAccount},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(listReply))
			require.NoError(t, err)
		case "POST":
			req := &orgv1.CreateServiceAccountRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			serviceAccount := &orgv1.User{
				Id:                 55555,
				ServiceName:        req.User.ServiceName,
				ServiceDescription: req.User.ServiceDescription,
			}
			createReply, err := utilv1.MarshalJSONToBytes(&orgv1.CreateServiceAccountReply{
				Error: nil,
				User:  serviceAccount,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(createReply))
			require.NoError(t, err)
		case "PUT":
			req := &orgv1.UpdateServiceAccountRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			updateReply, err := utilv1.MarshalJSONToBytes(&orgv1.UpdateServiceAccountReply{
				Error: nil,
				User:  req.User,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(updateReply))
			require.NoError(t, err)
		case "DELETE":
			req := &orgv1.DeleteServiceAccountRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			updateReply, err := utilv1.MarshalJSONToBytes(&orgv1.DeleteServiceAccountReply{
				Error: nil,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(updateReply))
			require.NoError(t, err)
		}
	}
}

var (
	keyStore        = map[int32]*schedv1.ApiKey{}
	keyIndex        = int32(1)
	keyTimestamp, _ = types.TimestampProto(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
)

func init() {
	fillKeyStore()
}
// Handler for: "/api/api_keys"
func (c *CloudRouter) HandleApiKeys(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			req := &schedv1.CreateApiKeyRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.ApiKey.AccountId)
			apiKey := req.ApiKey
			apiKey.Id = keyIndex
			apiKey.Key = fmt.Sprintf("MYKEY%d", keyIndex)
			apiKey.Secret = fmt.Sprintf("MYSECRET%d", keyIndex)
			apiKey.Created = keyTimestamp
			if req.ApiKey.UserId == 0 {
				apiKey.UserId = 23
			} else {
				apiKey.UserId = req.ApiKey.UserId
			}
			keyIndex++
			keyStore[apiKey.Id] = apiKey
			b, err := utilv1.MarshalJSONToBytes(&schedv1.CreateApiKeyReply{ApiKey: apiKey})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == "GET" {
			require.NotEmpty(t, r.URL.Query().Get("account_id"))
			apiKeys := apiKeysFilter(r.URL)
			// Return sorted data or the test output will not be stable
			sort.Sort(ApiKeyList(apiKeys))
			b, err := utilv1.MarshalJSONToBytes(&schedv1.GetApiKeysReply{ApiKeys: apiKeys})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/api_keys/{key}"
func (c *CloudRouter) HandleApiKey(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		keyStr := vars["key"]
		keyId, err := strconv.Atoi(keyStr)
		require.NoError(t, err)
		index := int32(keyId)
		apiKey := keyStore[index]
		if r.Method == "PUT" {
			req := &schedv1.UpdateApiKeyRequest{}
			err = utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			apiKey.Description = req.ApiKey.Description
			result := &schedv1.UpdateApiKeyReply{
				ApiKey: apiKey,
				Error:  nil,
			}
			b, err := utilv1.MarshalJSONToBytes(result)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == "DELETE" {
			req := &schedv1.DeleteApiKeyRequest{}
			err = utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			delete(keyStore, index)
			result := &schedv1.DeleteApiKeyReply{
				ApiKey: apiKey,
				Error:  nil,
			}
			b, err := utilv1.MarshalJSONToBytes(result)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/clusters"
func (c *CloudRouter) HandleClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "POST" {
			c.HandleKafkaClusterCreate(t)(w, r)
		} else if r.Method == "GET" {
			cluster := schedv1.KafkaCluster{
				Id:              "lkc-123",
				Name:            "abc",
				Deployment:      &schedv1.Deployment{Sku: productv1.Sku_BASIC},
				Durability:      0,
				Status:          0,
				Region:          "us-central1",
				ServiceProvider: "gcp",
			}
			clusterMultizone := schedv1.KafkaCluster{
				Id:              "lkc-456",
				Name:            "def",
				Deployment:      &schedv1.Deployment{Sku: productv1.Sku_BASIC},
				Durability:      1,
				Status:          0,
				Region:          "us-central1",
				ServiceProvider: "gcp",
			}
			b, err := utilv1.MarshalJSONToBytes(&schedv1.GetKafkaClustersReply{
				Clusters: []*schedv1.KafkaCluster{&cluster, &clusterMultizone},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}
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
// Handler for GET "/api/clusters/lkc-dedicated"
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

// Handler for GET "/api/clusters/lkc-dedicated-pending"
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

// Handler for GET "/api/clusters/lkc-dedicated-with-encryption"
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
// Handler for: "api/env_metadata"
func (c *CloudRouter) HandleEnvMetadata(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		clouds := []*schedv1.CloudMetadata{
			{
				Id:   "gcp",
				Name: "Google Cloud Platform",
				Regions: []*schedv1.Region{
					{
						Id:            "asia-southeast1",
						Name:          "asia-southeast1 (Singapore)",
						IsSchedulable: true,
					},
					{
						Id:            "asia-east2",
						Name:          "asia-east2 (Hong Kong)",
						IsSchedulable: true,
					},
				},
			},
			{
				Id:   "aws",
				Name: "Amazon Web Services",
				Regions: []*schedv1.Region{
					{
						Id:            "ap-northeast-1",
						Name:          "ap-northeast-1 (Tokyo)",
						IsSchedulable: false,
					},
					{
						Id:            "us-east-1",
						Name:          "us-east-1 (N. Virginia)",
						IsSchedulable: true,
					},
				},
			},
			{
				Id:   "azure",
				Name: "Azure",
				Regions: []*schedv1.Region{
					{
						Id:            "southeastasia",
						Name:          "southeastasia (Singapore)",
						IsSchedulable: false,
					},
				},
			},
		}
		reply, err := utilv1.MarshalJSONToBytes(&schedv1.GetEnvironmentMetadataReply{
			Clouds: clouds,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}
// Handler for: "/api/schema_registries"
func (c *CloudRouter) HandleSchemaRegistriesRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		q := r.URL.Query()
		id := q.Get("id")
		if id == "" {
			id = "lsrc-1234"
		}
		accountId := q.Get("account_id")
		srCluster := &schedv1.SchemaRegistryCluster{
			Id:        id,
			AccountId: accountId,
			Name:      "account schema-registry",
			Endpoint:  "SASL_SSL://sr-endpoint",
		}
		fmt.Println(srCluster)
		b, err := utilv1.MarshalJSONToBytes(&schedv1.GetSchemaRegistryClusterReply{
			Cluster: srCluster,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
// Handler for: "/api/ksqls"
func (c *CloudRouter) HandleKsqlsRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ksqlCluster1 := &schedv1.KSQLCluster{
			Id:                "lksqlc-ksql5",
			AccountId:         "25",
			KafkaClusterId:    "lkc-qwert",
			OutputTopicPrefix: "pksqlc-abcde",
			Name:              "account ksql",
			Storage:           101,
			Endpoint:          "SASL_SSL://ksql-endpoint",
		}
		ksqlCluster2 := &schedv1.KSQLCluster{
			Id:                "lksqlc-woooo",
			AccountId:         "25",
			KafkaClusterId:    "lkc-zxcvb",
			OutputTopicPrefix: "pksqlc-ghjkl",
			Name:              "kay cee queue elle",
			Storage:           123,
			Endpoint:          "SASL_SSL://ksql-endpoint",
		}
		if r.Method == "POST" {
			reply, err := utilv1.MarshalJSONToBytes(&schedv1.GetKSQLClusterReply{
				Cluster: ksqlCluster1,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(reply))
			require.NoError(t, err)
		} else if r.Method == "GET" {
			listReply, err := utilv1.MarshalJSONToBytes(&schedv1.GetKSQLClustersReply{
				Clusters: []*schedv1.KSQLCluster{ksqlCluster1, ksqlCluster2},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(listReply))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/ksqls/{id}"
func (c *CloudRouter) HandleKsqlRequests(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ksqlId := vars["id"]
		switch ksqlId {
		case "lksqlc-ksql1":
			ksqlCluster := &schedv1.KSQLCluster{
				Id:                "lksqlc-ksql1",
				AccountId:         "25",
				KafkaClusterId:    "lkc-12345",
				OutputTopicPrefix: "pksqlc-abcde",
				Name:              "account ksql",
				Storage:           101,
				Endpoint:          "SASL_SSL://ksql-endpoint",
			}
			reply, err := utilv1.MarshalJSONToBytes(&schedv1.GetKSQLClusterReply{
				Cluster: ksqlCluster,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(reply))
			require.NoError(t, err)
		case "lksqlc-12345":
			ksqlCluster := &schedv1.KSQLCluster{
				Id:                "lksqlc-12345",
				AccountId:         "25",
				KafkaClusterId:    "lkc-abcde",
				OutputTopicPrefix: "pksqlc-zxcvb",
				Name:              "account ksql",
				Storage:           130,
				Endpoint:          "SASL_SSL://ksql-endpoint",
			}
			reply, err := utilv1.MarshalJSONToBytes(&schedv1.GetKSQLClusterReply{
				Cluster: ksqlCluster,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(reply))
			require.NoError(t, err)
		default:
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/users"
func (c *CloudRouter) HandleUsers(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			users := []*orgv1.User{
				buildUser(1, "bstrauch@confluent.io", "Brian", "Strauch", "u11"),
				buildUser(2, "mtodzo@confluent.io", "Miles", "Todzo", "u-17"),
				buildUser(3, "u-11aaa@confluent.io", "11", "Aaa", "u-11aaa"),
				buildUser(4, "u-22bbb@confluent.io", "22", "Bbb", "u-22bbb"),
				buildUser(5, "u-33ccc@confluent.io", "33", "Ccc", "u-33ccc"),
			}
			userId := r.URL.Query().Get("id")
			if userId != "" {
				intId, err := strconv.Atoi(userId)
				require.NoError(t, err)
				if int32(intId) == deactivatedUserID {
					users = []*orgv1.User{}
				}
			}
			res := orgv1.GetUsersReply{
				Users:                users,
				Error:                nil,
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/users/{id}
func (c *CloudRouter) HandleUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		res := orgv1.DeleteUserReply{
			Error: nil,
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}
// Handler for: "/api/organizations/{id}/invites"
func (c *CloudRouter) HandleInvite(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bs := string(body)
		if strings.Contains(bs, "test@error.io") {
			w.WriteHeader(http.StatusBadRequest)
		} else {
			res := flowv1.SendInviteReply{
				Error: nil,
				User: buildUser(1, "miles@confluent.io", "Miles", "Todzo", ""),
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}"
func (c *CloudRouter) HandleConnector(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
//Handler for: ""/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/pause"
func (c *CloudRouter) HandleConnectorPause(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
//Handler for: ""/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/resume"
func (c *CloudRouter) HandleConnectorResume(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors"
func (c *CloudRouter) HandleConnectors(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["env"]
		clusterId := vars["cluster"]
		if r.Method == "GET" {
			connectorExpansion := &opv1.ConnectorExpansion{
				Id: &opv1.ConnectorId{Id: "lcc-123"},
				Info: &opv1.ConnectorInfo{
					Name:   "az-connector",
					Type:   "Sink",
					Config: map[string]string{},
				},
				Status: &opv1.ConnectorStateInfo{Name: "az-connector", Connector: &opv1.ConnectorState{State: "Running"},
					Tasks: []*opv1.TaskState{{Id: 1, State: "Running"}},
				}}
			listReply, err := json.Marshal(map[string]*opv1.ConnectorExpansion{"lcc-123": connectorExpansion})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(listReply))
			require.NoError(t, err)
		} else if r.Method == "POST" {
			var request opv1.ConnectorInfo
			err := utilv1.UnmarshalJSON(r.Body, &request)
			require.NoError(t, err)
			connector1 := &schedv1.Connector{
				Name:           request.Name,
				KafkaClusterId: clusterId,
				AccountId:      envId,
				UserConfigs:    request.Config,
				Plugin:         request.Config["connector.class"],
			}
			reply, err := utilv1.MarshalJSONToBytes(connector1)
			require.NoError(t, err)
			_, err = io.WriteString(w, string(reply))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors-plugins"
func (c *CloudRouter) HandlePlugins(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			connectorPlugin1 := &opv1.ConnectorPluginInfo{
				Class: "AzureBlobSink",
				Type:  "Sink",
			}
			connectorPlugin2 := &opv1.ConnectorPluginInfo{
				Class: "GcsSink",
				Type:  "Sink",
			}
			listReply, err := json.Marshal([]*opv1.ConnectorPluginInfo{connectorPlugin1, connectorPlugin2})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(listReply))
			require.NoError(t, err)
		}
	}
}
// Handler for: "/api/accounts/{env}/clusters/{cluster}/connector-plugins/{plugin}/config/validate"
func (c *CloudRouter) HandleConnectCatalog(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		configInfos := &opv1.ConfigInfos{
			Name:       "",
			Groups:     nil,
			ErrorCount: 1,
			Configs: []*opv1.Configs{
				{
					Value: &opv1.ConfigValue{
						Name:   "kafka.api.key",
						Errors: []string{"\"kafka.api.key\" is required"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "kafka.api.secret",
						Errors: []string{"\"kafka.api.secret\" is required"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "topics",
						Errors: []string{"\"topics\" is required"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "data.format",
						Errors: []string{"\"data.format\" is required", "Value \"null\" doesn't belong to the property's \"data.format\" enum"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "gcs.credentials.config",
						Errors: []string{"\"gcs.credentials.config\" is required"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "gcs.bucket.name",
						Errors: []string{"\"gcs.bucket.name\" is required"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "time.interval",
						Errors: []string{"\"data.format\" is required", "Value \"null\" doesn't belong to the property's \"time.interval\" enum"},
					},
				},
				{
					Value: &opv1.ConfigValue{
						Name:   "tasks.max",
						Errors: []string{"\"tasks.max\" is required"},
					},
				},
			},
		}
		reply, err := json.Marshal(configInfos)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

func (c *CloudRouter) HandleConnectUpdate(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
		return
	}
}
// Handler for: "/api/metadata/security/v2alpha1/authenticate"
func (c CloudRouter) HandleV2Authenticate(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		reply := &mds.AuthenticationResponse{
			AuthToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			TokenType: "dunno",
			ExpiresIn: 9999999999,
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}
