package testserver

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"gopkg.in/launchdarkly/go-sdk-common.v2/lduser"

	corev1 "github.com/confluentinc/cc-structs/kafka/core/v1"
	flowv1 "github.com/confluentinc/cc-structs/kafka/flow/v1"
	orgv1 "github.com/confluentinc/cc-structs/kafka/org/v1"
	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"

	"github.com/confluentinc/cli/internal/pkg/ccstructs"
	"github.com/confluentinc/cli/internal/pkg/errors"
)

var (
	environments       = []*ccloudv1.Account{{Id: "a-595", Name: "default"}, {Id: "not-595", Name: "other"}, {Id: "env-123", Name: "env123"}, {Id: SRApiEnvId, Name: "srUpdate"}}
	keyIndex           = int32(3)
	resourceIdMap      = map[int32]string{auditLogServiceAccountID: auditLogServiceAccountResourceID, serviceAccountID: serviceAccountResourceID}

	RegularOrg = &ccloudv1.Organization{
		Id:   321,
		Name: "test-org",
	}
	SuspendedOrg = func(eventType ccloudv1.SuspensionEventType) *ccloudv1.Organization {
		return &ccloudv1.Organization{
			Id:   321,
			Name: "test-org",
			SuspensionStatus: &ccloudv1.SuspensionStatus{
				Status:    ccloudv1.SuspensionStatusType_SUSPENSION_COMPLETED,
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

	serviceAccountID           = int32(12345)
	serviceAccountResourceID   = "sa-12345"
	identityProviderResourceID = "op-12345"
	identityPoolResourceID     = "pool-12345"
	deactivatedUserID          = int32(6666)
	deactivatedResourceID      = "sa-6666"

	auditLogServiceAccountID         = int32(1337)
	auditLogServiceAccountResourceID = "sa-1337"

	PromoTestCode = "PromoTestCode"

	exampleSRPriceKey   = "aws:us-west-2:free:1:max"
	exampleSRPriceTable = "SchemaRegistry"
	exampleSRPriceUnit  = "Schema-Hour"
	exampleSchemaLimit  = 1000
)

// Handler for: "/api/me"
func (c *CloudRouter) HandleMe(t *testing.T, isAuditLogEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		orgResourceId := os.Getenv("CONFLUENT_CLOUD_ORGANIZATION_ID")
		if orgResourceId == "" {
			orgResourceId = "abc-123"
		}

		org := &ccloudv1.Organization{Id: 42, ResourceId: orgResourceId, Name: "Confluent"}
		if !isAuditLogEnabled {
			org.AuditLog = &ccloudv1.AuditLog{
				ClusterId:        "lkc-ab123",
				AccountId:        "env-987zy",
				ServiceAccountId: auditLogServiceAccountID,
				TopicName:        "confluent-audit-log-events",
			}
		}

		b, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetMeReply{
			User: &ccloudv1.User{
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
		req := new(ccloudv1.AuthenticateRequest)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		res := new(ccloudv1.AuthenticateReply)

		switch req.Email {
		case "incorrect@user.com":
			w.WriteHeader(http.StatusForbidden)
		case "suspended@user.com":
			w.WriteHeader(http.StatusForbidden)
			res.Error = &ccloudv1.Error{Message: errors.SuspendedOrganizationSuggestions}
		case "end-of-free-trial-suspended@user.com":
			res.Token = "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE"
			res.Organization = SuspendedOrg(ccloudv1.SuspensionEventType_SUSPENSION_EVENT_END_OF_FREE_TRIAL)
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

		res := &ccloudv1.GetLoginRealmReply{
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
				b, err := ccstructs.MarshalJSONToBytes(&ccloudv1.GetAccountReply{Account: env})
				require.NoError(t, err)
				_, err = io.WriteString(w, string(b))
				require.NoError(t, err)
			case http.MethodPut: // called by `environment create`
				req := &ccloudv1.CreateAccountRequest{}
				err := ccstructs.UnmarshalJSON(r.Body, req)
				require.NoError(t, err)
				env.Name = req.Account.Name
				b, err := ccstructs.MarshalJSONToBytes(&ccloudv1.CreateAccountReply{Account: env})
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
			req := &ccloudv1.CreateAccountRequest{}
			err := ccstructs.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			account := &ccloudv1.Account{
				Id:             "a-5555",
				Name:           req.Account.Name,
				OrganizationId: 0,
			}
			b, err := ccstructs.MarshalJSONToBytes(&ccloudv1.CreateAccountReply{
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
			req := &ccloudv1.UpdatePaymentInfoRequest{}
			err := ccstructs.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.StripeToken)

			res := &ccloudv1.UpdatePaymentInfoReply{}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet: // admin payment describe
			res := ccloudv1.GetPaymentInfoReply{
				Card: &ccloudv1.Card{
					Cardholder: "Miles Todzo",
					Brand:      "Visa",
					Last4:      "4242",
					ExpMonth:   "01",
					ExpYear:    "99",
				},
				Organization: &ccloudv1.Organization{Id: 0},
			}
			data, err := json.Marshal(res)
			require.NoError(t, err)
			_, err = w.Write(data)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/service_accounts"
func (c *CloudRouter) HandleServiceAccounts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			res := &orgv1.GetServiceAccountsReply{
				Users: []*orgv1.User{
					{
						Id:                 serviceAccountID,
						ResourceId:         serviceAccountResourceID,
						ServiceName:        "service_account",
						ServiceDescription: "at your service.",
					},
					{
						Id:                 1,
						ResourceId:         "sa-00001",
						ServiceName:        "KSQL.lksqlc-12345",
						ServiceDescription: "ksqlDB service account",
					},
				},
			}
			listReply, err := ccstructs.MarshalJSONToBytes(res)
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

// Handler for: "api/env_metadata"
func (c *CloudRouter) HandleEnvMetadata(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clouds := []*ccloudv1.CloudMetadata{
			{
				Id:   "gcp",
				Name: "Google Cloud Platform",
				Regions: []*ccloudv1.Region{
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
				Regions: []*ccloudv1.Region{
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
				Regions: []*ccloudv1.Region{
					{
						Id:            "southeastasia",
						Name:          "southeastasia (Singapore)",
						IsSchedulable: false,
					},
				},
			},
		}
		reply, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetEnvironmentMetadataReply{
			Clouds: clouds,
		})
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
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
		b, err := ccstructs.MarshalJSONToBytes(&res)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(b))
		require.NoError(t, err)
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
		req := &ccloudv1.SignupRequest{}
		err := ccstructs.UnmarshalJSON(r.Body, req)
		require.NoError(t, err)
		require.NotEmpty(t, req.Organization.Name)
		require.NotEmpty(t, req.User)
		require.NotEmpty(t, req.Credentials)
		signupReply := &ccloudv1.SignupReply{Organization: &ccloudv1.Organization{}}
		reply, err := ccstructs.MarshalJSONToBytes(signupReply)
		require.NoError(t, err)
		_, err = io.WriteString(w, string(reply))
		require.NoError(t, err)
	}
}

// Handler for: "/ldapi/sdk/eval/{env}/users/{user}"
func (c *CloudRouter) HandleLaunchDarkly(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ldUserData, err := base64.StdEncoding.DecodeString(vars["user"])
		require.NoError(t, err)

		ldUser := lduser.User{}
		require.NoError(t, json.Unmarshal(ldUserData, &ldUser))

		w.Header().Set("Content-Type", "application/json")
		flags := map[string]interface{}{
			"testBool":                               true,
			"testString":                             "string",
			"testInt":                                1,
			"testJson":                               map[string]interface{}{"key": "val"},
			"cli.deprecation_notices":                []map[string]interface{}{},
			"cli.client_quotas.enable":               true,
			"cli.stream_designer.source_code.enable": true,
		}

		val, ok := ldUser.GetCustom("org.resource_id")
		if ok && val.StringValue() == "multicluster-key-org" {
			flags["cli.multicluster-api-keys.enable"] = true
		}

		err = json.NewEncoder(w).Encode(&flags)
		require.NoError(t, err)
	}
}

// Handler for: "/api/external_identities"
func handleExternalIdentities(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		res := &ccloudv1.CreateExternalIdentityResponse{IdentityName: "id-xyz"}
		err := json.NewEncoder(w).Encode(res)
		require.NoError(t, err)
	}
}
