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

	ccloudv1 "github.com/confluentinc/ccloud-sdk-go-v1-public"
	"github.com/confluentinc/mds-sdk-go-public/mdsv1"

	"github.com/confluentinc/cli/v4/pkg/ccstructs"
	"github.com/confluentinc/cli/v4/pkg/errors"
)

var (
	environments = []*ccloudv1.Account{
		{Id: "env-596", Name: "default", OrgResourceId: "abc-123"},
		{Id: "env-595", Name: "other"},
		{Id: "env-123", Name: "env123"},
		{Id: SRApiEnvId, Name: "srUpdate"},
		{Id: "env-987zy", Name: "confluent-audit-log"},
	}
	keyIndex      = int32(3)
	resourceIdMap = map[int32]string{auditLogServiceAccountId: auditLogServiceAccountResourceId, serviceAccountId: serviceAccountResourceId}

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
	serviceAccountId                 = int32(12345)
	serviceAccountResourceId         = "sa-12345"
	groupMappingId                   = "group-abc"
	identityProviderId               = "op-12345"
	identityPoolId                   = "pool-12345"
	ipGroupId                        = "ipg-wjnde"
	ipFilterId                       = "ipf-34dq3"
	deactivatedUserId                = int32(6666)
	deactivatedUserResourceId        = "sa-6666"
	auditLogServiceAccountId         = int32(1337)
	auditLogServiceAccountResourceId = "sa-1337"
)

// Handler for: "/api/me"
func handleMe(t *testing.T, isAuditLogEnabled bool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		organizationId := os.Getenv("CONFLUENT_CLOUD_ORGANIZATION_ID")
		if organizationId == "" {
			organizationId = "abc-123"
		}

		org := &ccloudv1.Organization{
			Id:         42,
			ResourceId: organizationId,
			Name:       "Confluent",
		}
		if isAuditLogEnabled {
			org.AuditLog = &ccloudv1.AuditLog{
				ClusterId:                "lkc-ab123",
				AccountId:                "env-987zy",
				ServiceAccountId:         auditLogServiceAccountId,
				ServiceAccountResourceId: "sa-1337",
				TopicName:                "confluent-audit-log-events",
			}
		}

		if os.Getenv("IS_ORG_ON_MARKETPLACE") == "true" {
			org.Marketplace = &ccloudv1.Marketplace{Partner: ccloudv1.MarketplacePartner_AWS}
		}

		environmentList := environments
		if os.Getenv("CONFLUENT_CLOUD_EMAIL") == "no-environment-user@example.com" {
			environmentList = []*ccloudv1.Account{}
		}
		b, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetMeReply{
			User: &ccloudv1.User{
				Id:         23,
				Email:      "mhe@confluent.io",
				FirstName:  "Muwei",
				ResourceId: "u-44ddd",
			},
			Accounts:     environmentList,
			Organization: org,
		})
		require.NoError(t, err)
		_, err = io.Writer.Write(w, b)
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
			res.Token = "eyJ.eyJ.malformed"
		case "invalid@user.com":
			res.Token = "eyJ.eyJ.invalid"
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

// Handler for: "/api/organizations/{id}/payment_info"
func handlePaymentInfo(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost: // billing payment update
			req := &ccloudv1.UpdatePaymentInfoRequest{}
			err := ccstructs.UnmarshalJSON(r.Body, req)
			require.NoError(t, err)
			require.NotEmpty(t, req.StripeToken)

			res := &ccloudv1.UpdatePaymentInfoReply{}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet: // billing payment describe
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
func handleServiceAccounts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			res := &ccloudv1.GetServiceAccountsReply{
				Users: []*ccloudv1.User{
					{
						Id:                 serviceAccountId,
						ResourceId:         serviceAccountResourceId,
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
			_, err = io.Writer.Write(w, listReply)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/api/service_accounts/{id}"
func handleServiceAccount(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		idStr := mux.Vars(r)["id"]
		id, err := strconv.ParseInt(idStr, 10, 32)
		require.NoError(t, err)
		userId := int32(id)
		switch r.Method {
		case http.MethodGet:
			res := &ccloudv1.GetServiceAccountReply{
				User: &ccloudv1.User{
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
func handleEnvMetadata(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		clouds := []*ccloudv1.CloudMetadata{
			{
				Id:   "gcp",
				Name: "Google Cloud Platform",
				Regions: []*ccloudv1.Region{
					{
						Id:   "asia-southeast1",
						Name: "asia-southeast1 (Singapore)",
						Schedulability: &ccloudv1.Schedulability{
							DedicatedNetwork: &ccloudv1.Schedulability_Tenancy{
								DedicatedCluster: &ccloudv1.Schedulability_Tenancy_Durability{
									High: []ccloudv1.NetworkType{
										ccloudv1.NetworkType_VPC_PEERING,
										ccloudv1.NetworkType_TRANSIT_GATEWAY,
										ccloudv1.NetworkType_PRIVATE_LINK,
									},
								},
							},
						},
					},
					{
						Id:   "asia-east2",
						Name: "asia-east2 (Hong Kong)",
						Schedulability: &ccloudv1.Schedulability{
							DedicatedNetwork: &ccloudv1.Schedulability_Tenancy{
								DedicatedCluster: &ccloudv1.Schedulability_Tenancy_Durability{
									High: []ccloudv1.NetworkType{
										ccloudv1.NetworkType_VPC_PEERING,
										ccloudv1.NetworkType_TRANSIT_GATEWAY,
										ccloudv1.NetworkType_PRIVATE_LINK,
									},
								},
							},
						},
					},
				},
			},
			{
				Id:   "aws",
				Name: "Amazon Web Services",
				Regions: []*ccloudv1.Region{
					{
						Id:   "ap-northeast-1",
						Name: "ap-northeast-1 (Tokyo)",
						Schedulability: &ccloudv1.Schedulability{
							DedicatedNetwork: &ccloudv1.Schedulability_Tenancy{
								DedicatedCluster: &ccloudv1.Schedulability_Tenancy_Durability{
									High: []ccloudv1.NetworkType{},
								},
							},
						},
					},
					{
						Id:   "us-east-1",
						Name: "us-east-1 (N. Virginia)",
						Schedulability: &ccloudv1.Schedulability{
							DedicatedNetwork: &ccloudv1.Schedulability_Tenancy{
								DedicatedCluster: &ccloudv1.Schedulability_Tenancy_Durability{
									High: []ccloudv1.NetworkType{
										ccloudv1.NetworkType_VPC_PEERING,
										ccloudv1.NetworkType_TRANSIT_GATEWAY,
										ccloudv1.NetworkType_PRIVATE_LINK,
									},
								},
							},
						},
					},
				},
			},
			{
				Id:   "azure",
				Name: "Azure",
				Regions: []*ccloudv1.Region{
					{
						Id:   "southeastasia",
						Name: "southeastasia (Singapore)",
						Schedulability: &ccloudv1.Schedulability{
							DedicatedNetwork: &ccloudv1.Schedulability_Tenancy{
								DedicatedCluster: &ccloudv1.Schedulability_Tenancy_Durability{
									High: []ccloudv1.NetworkType{},
								},
							},
						},
					},
				},
			},
		}
		reply, err := ccloudv1.MarshalJSONToBytes(&ccloudv1.GetEnvironmentMetadataReply{
			Clouds: clouds,
		})
		require.NoError(t, err)
		_, err = io.Writer.Write(w, reply)
		require.NoError(t, err)
	}
}

// Handler for: "/api/users"
func handleUsers(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users := []*ccloudv1.User{
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
				if int32(intId) == deactivatedUserId {
					users = []*ccloudv1.User{}
				}
			}
			res := ccloudv1.GetUsersReply{
				Users: users,
				Error: nil,
			}
			email := r.URL.Query().Get("email")
			if email != "" {
				for _, u := range users {
					if u.Email == email {
						res = ccloudv1.GetUsersReply{
							Users: []*ccloudv1.User{u},
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

// Handler for: "/api/metadata/security/v2alpha1/authenticate"
func handleV2Authenticate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		reply := &mdsv1.AuthenticationResponse{
			AuthToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			TokenType: "dunno",
			ExpiresIn: 9999999999,
		}
		b, err := json.Marshal(&reply)
		require.NoError(t, err)
		_, err = io.Writer.Write(w, b)
		require.NoError(t, err)
	}
}

// Handler for: "/ldapi/sdk/eval/{env}/users/{user}"
func handleLaunchDarkly(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ldUserData, err := base64.StdEncoding.DecodeString(vars["user"])
		require.NoError(t, err)

		ldUser := lduser.User{}
		require.NoError(t, json.Unmarshal(ldUserData, &ldUser))

		flags := map[string]any{
			"testBool":                                  true,
			"testString":                                "string",
			"testInt":                                   1,
			"testAnotherInt":                            99,
			"testJson":                                  map[string]any{"key": "val"},
			"cli.deprecation_notices":                   []map[string]any{},
			"cli.client_quotas.enable":                  true,
			"cli.stream_designer.source_code.enable":    true,
			"flink.rbac.namespace.cli.enable":           true,
			"auth.rbac.identity_admin.enable":           true,
			"flink.language_service.enable_diagnostics": true,
			"cloud_growth.marketplace_linking_advertisement_experiment.enable": true,
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
