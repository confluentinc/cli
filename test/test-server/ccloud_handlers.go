package test_server

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/proto"
	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	billingv1 "github.com/confluentinc/cc-structs/kafka/billing/v1"
	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	productv1 "github.com/confluentinc/cc-structs/kafka/product/core/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	opv1 "github.com/confluentinc/cc-structs/operator/v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	environments    = []*orgv1.Account{{Id: "a-595", Name: "default"}, {Id: "not-595", Name: "other"}, {Id: "env-123", Name: "env123"}, {Id: SRApiEnvId, Name: "srUpdate"}}
	keyStore        = map[int32]*schedv1.ApiKey{}
	keyIndex        = int32(1)
	keyTimestamp, _ = types.TimestampProto(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
	resourceIdMap   = map[int32]string{auditLogServiceAccountID: auditLogServiceAccountResourceID, serviceAccountID: serviceAccountResourceID}
)

const (
	exampleAvailability = "low"
	exampleCloud        = "aws"
	exampleClusterType  = "basic"
	exampleMetric       = "ConnectNumRecords"
	exampleNetworkType  = "internet"
	examplePrice        = 1
	exampleRegion       = "us-east-1"
	exampleUnit         = "GB"

	serviceAccountID         = int32(12345)
	serviceAccountResourceID = "sa-12345"
	deactivatedUserID        = int32(6666)
	deactivatedResourceID    = "sa-6666"

	auditLogServiceAccountID         = int32(1337)
	auditLogServiceAccountResourceID = "sa-1337"
)

// Fill API keyStore with default data
func init() {
	fillKeyStore()
}

// Handler for: "/api/me"
func (c *CloudRouter) HandleMe(t *testing.T, isAuditLogEnabled bool) func(http.ResponseWriter, *http.Request) {
	org := &orgv1.Organization{Id: 42, ResourceId: "abc-123", Name: "Confluent"}
	if !isAuditLogEnabled {
		org.AuditLog = &orgv1.AuditLog{
			ClusterId:        "lkc-ab123",
			AccountId:        "env-987zy",
			ServiceAccountId: auditLogServiceAccountID,
			TopicName:        "confluent-audit-log-events",
		}
	}
	return func(w http.ResponseWriter, r *http.Request) {
		b, err := utilv1.MarshalJSONToBytes(&flowv1.GetMeReply{
			User: &orgv1.User{
				Id:         23,
				Email:      "cody@confluent.i",
				FirstName:  "Cody",
				ResourceId: "u-11aaa",
			},
			Accounts:     environments,
			Organization: org,
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
		case "suspended@user.com":
			w.WriteHeader(http.StatusForbidden)
			e := &struct {
				Error corev1.Error `json:"error"`
			}{
				Error: corev1.Error{Message: errors.SuspendedOrganizationSuggestions},
			}
			err := json.NewEncoder(w).Encode(e)
			req.NoError(err)
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

// Handler for: "/api/login/realm"
func (c *CloudRouter) HandleLoginRealm(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := require.New(t)
		reply := &flowv1.GetLoginRealmReply{}
		b, err := utilv1.MarshalJSONToBytes(reply)
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	}
}

// Handler for: "/api/accounts/{id}"
func (c *CloudRouter) HandleEnvironment(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		if valid, env := isValidEnvironmentId(environments, envId); valid {
			switch r.Method {
			case http.MethodGet:
				b, err := utilv1.MarshalJSONToBytes(&orgv1.GetAccountReply{Account: env})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			case http.MethodPut:
				req := &orgv1.UpdateAccountRequest{}
				err := utilv1.UnmarshalJSON(r.Body, req)
				require.NoError(t, err)
				env.Name = req.Account.Name
				b, err := utilv1.MarshalJSONToBytes(&orgv1.UpdateAccountReply{Account: env})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			case http.MethodDelete:
				b, err := utilv1.MarshalJSONToBytes(&orgv1.DeleteAccountReply{})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			}
		} else {
			// env not found
			w.WriteHeader(http.StatusNotFound)
		}
	}
}

// Handler for: "/api/accounts"
func (c *CloudRouter) HandleEnvironments(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			b, err := utilv1.MarshalJSONToBytes(&orgv1.ListAccountsReply{Accounts: environments})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == http.MethodPost {
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

// Handler for: "/api/organizations/{id}/payment_info"
func (c *CloudRouter) HandlePaymentInfo(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost: //admin payment update
			req := &orgv1.UpdatePaymentInfoRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.StripeToken)

			res := &orgv1.UpdatePaymentInfoReply{}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet: // admin payment describe
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
}

// Handler for "/api/organizations/"
func (c *CloudRouter) HandlePriceTable(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		prices := map[string]float64{
			strings.Join([]string{exampleCloud, exampleRegion, exampleAvailability, exampleClusterType, exampleNetworkType}, ":"): examplePrice,
		}

		res := &billingv1.GetPriceTableReply{
			PriceTable: &billingv1.PriceTable{
				PriceTable: map[string]*billingv1.UnitPrices{
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

// Handler for: "/api/organizations/{id}/promo_code_claims"
func (c *CloudRouter) HandlePromoCodeClaims(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			var tenDollars int64 = 10 * 10000

			// The time is set to noon so that all time zones display the same local time
			date := time.Date(2021, time.June, 16, 12, 0, 0, 0, time.UTC)
			expiration := &types.Timestamp{Seconds: date.Unix()}

			res := &billingv1.GetPromoCodeClaimsReply{
				Claims: []*billingv1.PromoCodeClaim{
					{
						Code:                 "PROMOCODE1",
						Amount:               tenDollars,
						Balance:              tenDollars,
						CreditExpirationDate: expiration,
					},
					{
						Code:                 "PROMOCODE2",
						Balance:              tenDollars,
						Amount:               tenDollars,
						CreditExpirationDate: expiration,
					},
				},
			}

			listReply, err := utilv1.MarshalJSONToBytes(res)
			require.NoError(t, err)
			_, err = w.Write(listReply)
			require.NoError(t, err)
		case http.MethodPost:
			res := &billingv1.ClaimPromoCodeReply{}

			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/service_accounts"
func (c *CloudRouter) HandleServiceAccounts(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			serviceAccount := &orgv1.User{
				Id:                 serviceAccountID,
				ResourceId:         serviceAccountResourceID,
				ServiceName:        "service_account",
				ServiceDescription: "at your service.",
			}
			listReply, err := utilv1.MarshalJSONToBytes(&orgv1.GetServiceAccountsReply{
				Users: []*orgv1.User{serviceAccount},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(listReply))
			require.NoError(t, err)
		case http.MethodPost:
			req := &orgv1.CreateServiceAccountRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			serviceAccount := &orgv1.User{
				Id:                 55555,
				ResourceId:         "sa-55555",
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
		case http.MethodPut:
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
		case http.MethodDelete:
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

// Handler for: "/api/service_accounts/{id}"
func (c *CloudRouter) HandleServiceAccount(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.ParseInt(idStr, 10, 32)
		require.NoError(t, err)
		userId := int32(id)

		switch r.Method {
		case http.MethodGet:
			res := &orgv1.GetServiceAccountReply{
				User: &orgv1.User{
					Id:         userId,
					ResourceId: resourceIdMap[userId],
				},
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)

			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/api_keys"
func (c *CloudRouter) HandleApiKeys(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			req := &schedv1.CreateApiKeyRequest{}
			err := utilv1.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.ApiKey.AccountId)

			if req.ApiKey.UserResourceId == "sa-123456" {
				b, err := utilv1.MarshalJSONToBytes(&schedv1.CreateApiKeyReply{Error: &corev1.Error{Message: "service account is not valid"}})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			}

			apiKey := req.ApiKey
			apiKey.Id = keyIndex
			apiKey.Key = fmt.Sprintf("MYKEY%d", keyIndex)
			apiKey.Secret = fmt.Sprintf("MYSECRET%d", keyIndex)
			apiKey.Created = keyTimestamp
			if req.ApiKey.UserId == 0 {
				apiKey.UserId = 23
				apiKey.UserResourceId = "u-44ddd"
			} else {
				apiKey.UserId = req.ApiKey.UserId
			}
			keyIndex++
			keyStore[apiKey.Id] = apiKey
			b, err := utilv1.MarshalJSONToBytes(&schedv1.CreateApiKeyReply{ApiKey: apiKey})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == http.MethodGet {
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
		if r.Method == http.MethodPut {
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
		} else if r.Method == http.MethodDelete {
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

// Handler for: "/api/ksqls"
func (c *CloudRouter) HandleKsqls(t *testing.T) func(http.ResponseWriter, *http.Request) {
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
		if r.Method == http.MethodPost {
			reply, err := utilv1.MarshalJSONToBytes(&schedv1.GetKSQLClusterReply{
				Cluster: ksqlCluster1,
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(reply))
			require.NoError(t, err)
		} else if r.Method == http.MethodGet {
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
func (c *CloudRouter) HandleKsql(t *testing.T) func(http.ResponseWriter, *http.Request) {
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
		if r.Method == http.MethodGet {
			users := []*orgv1.User{
				buildUser(1, "bstrauch@confluent.io", "Brian", "Strauch", "u11"),
				buildUser(2, "mtodzo@confluent.io", "Miles", "Todzo", "u-17"),
				buildUser(3, "u-11aaa@confluent.io", "11", "Aaa", "u-11aaa"),
				buildUser(4, "u-22bbb@confluent.io", "22", "Bbb", "u-22bbb"),
				buildUser(5, "u-33ccc@confluent.io", "33", "Ccc", "u-33ccc"),
				buildUser(23, "mhe@confluent.io", "Muwei", "He", "u-44ddd"),
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
				Users: users,
				Error: nil,
			}
			email := r.URL.Query().Get("email")
			if email != "" {
				for _, u := range users {
					if u.Email == email {
						res = orgv1.GetUsersReply{
							Users: []*orgv1.User{u},
							Error: nil,
						}
						break
					}
				}
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/users/{id}"
func (c *CloudRouter) HandleUser(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["id"]
		var res orgv1.DeleteUserReply
		switch userId {
		case "u-1":
			res = orgv1.DeleteUserReply{
				Error: &corev1.Error{Message: "user not found"},
			}
		default:
			res = orgv1.DeleteUserReply{
				Error: nil,
			}
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

// Handler for: "/api/user_profiles/{id}"
func (c *CloudRouter) HandleUserProfiles(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["id"]
		var res flowv1.GetUserProfileReply
		users := []*orgv1.User{
			buildUser(1, "bstrauch@confluent.io", "Brian", "Strauch", "u11"),
			buildUser(2, "mtodzo@confluent.io", "Miles", "Todzo", "u-17"),
			buildUser(3, "u-11aaa@confluent.io", "11", "Aaa", "u-11aaa"),
			buildUser(4, "u-22bbb@confluent.io", "22", "Bbb", "u-22bbb"),
			buildUser(5, "u-33ccc@confluent.io", "33", "Ccc", "u-33ccc"),
			buildUser(23, "mhe@confluent.io", "Muwei", "He", "u-44ddd"),
		}
		var user *orgv1.User
		switch userId {
		case "u-0":
			res = flowv1.GetUserProfileReply{
				Error: &corev1.Error{Message: "user not found"},
			}
		case "u11":
			user = users[0]
		case "u-17":
			user = users[1]
		case "u-11aaa":
			user = users[2]
		case "u-22bbb":
			user = users[3]
		case "u-33ccc":
			user = users[4]
		case "u-44ddd":
			user = users[5]
		default:
			res = flowv1.GetUserProfileReply{
				User: &flowv1.UserProfile{
					Email:      "cody@confluent.io",
					FirstName:  "Cody",
					ResourceId: "u-11aaa",
					UserStatus: flowv1.UserStatus_USER_STATUS_UNVERIFIED,
				},
			}
		}
		if userId != "u-0" {
			authConfig := &orgv1.AuthConfig{
				AllowedAuthMethods: []orgv1.AuthMethod{orgv1.AuthMethod_AUTH_METHOD_USERNAME_PWD, orgv1.AuthMethod_AUTH_METHOD_SSO},
			}
			res = flowv1.GetUserProfileReply{
				User: &flowv1.UserProfile{
					Email:      user.Email,
					FirstName:  user.FirstName,
					LastName:   user.LastName,
					ResourceId: user.ResourceId,
					UserStatus: flowv1.UserStatus_USER_STATUS_UNVERIFIED,
					AuthConfig: authConfig,
				},
			}
		}
		b, err := utilv1.MarshalJSONToBytes(&res)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
	}
}

// Handler for: "/api/organizations/{id}/invites"
func (c *CloudRouter) HandleInvite(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		body, _ := ioutil.ReadAll(r.Body)
		bs := string(body)
		var res flowv1.SendInviteReply
		switch {
		case strings.Contains(bs, "user@exists.com"):
			res = flowv1.SendInviteReply{
				Error: &corev1.Error{Message: "User is already active"},
				User:  nil,
			}
		default:
			res = flowv1.SendInviteReply{
				Error: nil,
				User:  buildUser(1, "miles@confluent.io", "Miles", "Todzo", ""),
			}
		}
		data, err := json.Marshal(res)
		require.NoError(t, err)
		_, err = w.Write(data)
		require.NoError(t, err)
	}
}

// Handler for: "/api/invitations"
func (c *CloudRouter) HandleInvitations(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			b, err := utilv1.MarshalJSONToBytes(&flowv1.ListInvitationsByOrgReply{
				Invitations: []*orgv1.Invitation{
					buildInvitation("1", "u-11aaa@confluent.io", "u-11aaa", "VERIFIED"),
					buildInvitation("2", "u-22bbb@confluent.io", "u-22bbb", "SENT"),
				},
			})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		} else if r.Method == http.MethodPost {
			body, _ := ioutil.ReadAll(r.Body)
			bs := string(body)
			var res flowv1.CreateInvitationReply
			if strings.Contains(bs, "user@exists.com") {
				res = flowv1.CreateInvitationReply{
					Error: &corev1.Error{Message: "User is already active"},
				}
			} else {
				res = flowv1.CreateInvitationReply{
					Error:      nil,
					Invitation: buildInvitation("1", "miles@confluent.io", "user1", "SENT"),
				}
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}"
func (c *CloudRouter) HandleConnector() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: ""/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/pause"
func (c *CloudRouter) HandleConnectorPause() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: ""/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/resume"
func (c *CloudRouter) HandleConnectorResume() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors"
func (c *CloudRouter) HandleConnectors(t *testing.T) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["env"]
		clusterId := vars["cluster"]
		if r.Method == http.MethodGet {
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
		} else if r.Method == http.MethodPost {
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
		if r.Method == http.MethodGet {
			connectorPlugin1 := &opv1.ConnectorPluginInfo{
				Class: "GcsSink",
				Type:  "Sink",
			}
			connectorPlugin2 := &opv1.ConnectorPluginInfo{
				Class: "AzureBlobSink",
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

// Handler for: "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/config"
func (c *CloudRouter) HandleConnectUpdate() func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
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

// Handler for: "/api/signup"
func (c *CloudRouter) HandleSignup(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &orgv1.SignupRequest{}
		err := utilv1.UnmarshalJSON(r.Body, req)
		require.NoError(t, err)
		require.NotEmpty(t, req.Organization.Name)
		require.NotEmpty(t, req.User)
		require.NotEmpty(t, req.Credentials)
		signupReply := &orgv1.SignupReply{Organization: &orgv1.Organization{}}
		reply, err := utilv1.MarshalJSONToBytes(signupReply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

// Handler for: "/api/email_verifications"
func (c *CloudRouter) HandleSendVerificationEmail(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		req := &flowv1.CreateEmailVerificationRequest{}
		err := utilv1.UnmarshalJSON(r.Body, req)
		require.NoError(t, err)
		w.WriteHeader(http.StatusOK)
	}
}
