package testserver

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	billingv1 "github.com/confluentinc/cc-structs/kafka/billing/v1"
	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	schedv1 "github.com/confluentinc/cc-structs/kafka/scheduler/v1"
	utilv1 "github.com/confluentinc/cc-structs/kafka/util/v1"
	bucketv1 "github.com/confluentinc/cire-bucket-service/protos/bucket/v1"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/resource"
)

var (
	environments       = []*orgv1.Account{{Id: "a-595", Name: "default"}, {Id: "not-595", Name: "other"}, {Id: "env-123", Name: "env123"}, {Id: SRApiEnvId, Name: "srUpdate"}}
	keyStore           = map[int32]*schedv1.ApiKey{}
	keyIndex           = int32(1)
	keyTimestamp, _    = types.TimestampProto(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
	resourceIdMap      = map[int32]string{auditLogServiceAccountID: auditLogServiceAccountResourceID, serviceAccountID: serviceAccountResourceID}
	resourceTypeToKind = map[string]string{resource.Kafka: "Cluster", resource.Ksql: "ksqlDB", resource.SchemaRegistry: "SchemaRegistry", resource.Cloud: "Cloud"}

	RegularOrg = &orgv1.Organization{
		Id:   321,
		Name: "test-org",
	}
	SuspendedOrg = func(eventType orgv1.SuspensionEventType) *orgv1.Organization {
		return &orgv1.Organization{
			Id:   321,
			Name: "test-org",
			SuspensionStatus: &orgv1.SuspensionStatus{
				Status:    orgv1.SuspensionStatusType_SUSPENSION_COMPLETED,
				EventType: eventType,
			},
		}
	}
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
func (c *CloudRouter) HandleMe(t *testing.T, isAuditLogEnabled bool) http.HandlerFunc {
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
				Email:      "cody@confluent.io",
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
func handleLogin(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(flowv1.AuthenticateRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		res := new(flowv1.AuthenticateReply)

		switch req.Email {
		case "incorrect@user.com":
			w.WriteHeader(http.StatusForbidden)
		case "suspended@user.com":
			w.WriteHeader(http.StatusForbidden)
			res.Error = &corev1.Error{Message: errors.SuspendedOrganizationSuggestions}
		case "end-of-free-trial-suspended@user.com":
			res.Token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
			res.Organization = SuspendedOrg(orgv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL)
		case "expired@user.com":
			res.Token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1MzAxMjQ4NTcsImV4cCI6MTUzMDAzODQ1NywiYXVkIjoid3d3LmV4YW1wbGUuY29tIiwic3ViIjoianJvY2tldEBleGFtcGxlLmNvbSJ9.Y2ui08GPxxuV9edXUBq-JKr1VPpMSnhjSFySczCby7Y"
		case "malformed@user.com":
			res.Token = "malformed"
		case "invalid@user.com":
			res.Token = "invalid"
		default:
			res.Token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
			res.Organization = RegularOrg
		}

		err = json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}

// Handler for: "/api/login/realm"
func handleLoginRealm(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		email := r.URL.Query().Get("email")

		res := &flowv1.GetLoginRealmReply{
			IsSso: strings.Contains(email, "sso"),
			Realm: "realm",
		}
		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}

// Handler for: "/api/accounts/{id}"
func (c *CloudRouter) HandleEnvironment(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		envId := vars["id"]
		if env := isValidEnvironmentId(environments, envId); env != nil {
			switch r.Method {
			case http.MethodGet: // called by `environment use`
				b, err := utilv1.MarshalJSONToBytes(&orgv1.GetAccountReply{Account: env})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			case http.MethodPut: // called by `environment create`
				req := &orgv1.UpdateAccountRequest{}
				err := utilv1.UnmarshalJSON(r.Body, req)
				require.NoError(t, err)
				env.Name = req.Account.Name
				b, err := utilv1.MarshalJSONToBytes(&orgv1.UpdateAccountReply{Account: env})
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

// Handler for: "/api/accounts" Post
func (c *CloudRouter) HandleEnvironments(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
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
func (c *CloudRouter) HandlePaymentInfo(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandlePriceTable(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandlePromoCodeClaims(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleServiceAccounts(t *testing.T) http.HandlerFunc {
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
		}
	}
}

// Handler for: "/api/service_accounts/{id}"
func (c *CloudRouter) HandleServiceAccount(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleApiKeys(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
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
				apiKey.UserResourceId = "u-44ddd"
			} else {
				apiKey.UserId = req.ApiKey.UserId
			}
			keyIndex++
			keyStore[apiKey.Id] = apiKey
			v2ApiKey := getV2ApiKey(apiKey)
			keyStoreV2[*v2ApiKey.Id] = v2ApiKey
			b, err := utilv1.MarshalJSONToBytes(&schedv1.CreateApiKeyReply{ApiKey: apiKey})
			require.NoError(t, err)
			_, err = io.WriteString(w, string(b))
			require.NoError(t, err)
		}
	}
}

// Handler for: "api/env_metadata"
func (c *CloudRouter) HandleEnvMetadata(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleKsqls(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleKsql(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleUsers(t *testing.T) http.HandlerFunc {
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

// Handler for: "/api/user_profiles/{id}"
func (c *CloudRouter) HandleUserProfiles(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleInvite(t *testing.T) http.HandlerFunc {
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
func (c *CloudRouter) HandleInvitations(t *testing.T) http.HandlerFunc {
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

// Handler for: "/api/metadata/security/v2alpha1/authenticate"
func (c CloudRouter) HandleV2Authenticate(t *testing.T) http.HandlerFunc {
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

// Handler for: "/ldapi/sdk/eval/{env}/users/{user}"
func (c *CloudRouter) HandleLaunchDarkly(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		jsonVal := map[string]interface{}{"key": "val"}
		flags := map[string]interface{}{"testBool": true, "testString": "string", "testInt": 1, "testJson": jsonVal}
		err := json.NewEncoder(w).Encode(&flags)
		require.NoError(t, err)
	}
}

// Handler for: "/api/external_identities"
func handleExternalIdentities(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := &bucketv1.CreateExternalIdentityResponse{IdentityName: "id-xyz"}
		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}
