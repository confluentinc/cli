package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var (
	keyStoreV2       = map[string]*apikeysv2.IamV2ApiKey{}
	keyTime          = apikeysv2.PtrTime(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
	roleBindingStore = []mdsv2.IamV2RoleBinding{
		buildRoleBinding("pool-12345", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123/identity-provider="+identityProviderResourceID),
		buildRoleBinding("u-11aaa", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("sa-12345", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("u-11aaa", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("u-22bbb", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("u-22bbb", "EnvironmentAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=a-595"),
		buildRoleBinding("u-33ccc", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("u-44ddd", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/group=readers"),
		buildRoleBinding("u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/topic=clicks-*"),
		buildRoleBinding("u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/topic=payroll"),
		buildRoleBinding("u-66fff", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/cloud-cluster=lkc-1111aaa/ksql=ksql-cluster-name-2222bbb"),
		buildRoleBinding("u-77ggg", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=a-595/schema-registry=lsrc-3333ccc/subject=clicks"),
	}
)

func init() {
	fillKeyStoreV2()
}

// Handler for: "/iam/v2/api-keys/{id}"
func handleIamApiKey(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		vars := mux.Vars(r)
		keyStr := vars["id"]
		switch r.Method {
		case http.MethodPatch:
			handleIamApiKeyUpdate(t, keyStr)(w, r)
		case http.MethodGet:
			handleIamApiKeyGet(t, keyStr)(w, r)
		case http.MethodDelete:
			handleIamApiKeyDelete(t, keyStr)(w, r)
		}
	}
}

func handleIamApiKeyUpdate(t *testing.T, keyStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(apikeysv2.IamV2ApiKey)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)
		apiKey := keyStoreV2[keyStr]
		apiKey.Spec.Description = req.Spec.Description
		err = json.NewEncoder(w).Encode(apiKey)
		require.NoError(t, err)
	}
}

func handleIamApiKeyGet(t *testing.T, keyStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if keyStr == "UNKNOWN" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		apiKey := keyStoreV2[keyStr]
		err := json.NewEncoder(w).Encode(apiKey)
		require.NoError(t, err)
	}
}

func handleIamApiKeyDelete(t *testing.T, keyStr string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if keyStr == "UNKNOWN" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		delete(keyStoreV2, keyStr)
		w.WriteHeader(http.StatusNoContent)
	}
}

// Handler for: "/iam/v2/api-keys"
func handleIamApiKeys(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.Method == http.MethodPost {
			handleIamApiKeysCreate(t)(w, r)
		} else if r.Method == http.MethodGet {
			apiKeysList := apiKeysFilterV2(r.URL)
			err := json.NewEncoder(w).Encode(apiKeysList)
			require.NoError(t, err)
		}
	}
}

func handleIamApiKeysCreate(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		req := new(apikeysv2.IamV2ApiKey)
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)
		if req.Spec.Owner.Id == "sa-123456" {
			err = writeServiceAccountInvalidError(w)
			require.NoError(t, err)
			return
		}
		apiKey := req
		apiKey.Id = apikeysv2.PtrString(fmt.Sprintf("MYKEY%d", keyIndex))
		apiKey.Spec = &apikeysv2.IamV2ApiKeySpec{
			Owner:       req.Spec.Owner,
			Secret:      apikeysv2.PtrString(fmt.Sprintf("MYSECRET%d", keyIndex)),
			Resource:    req.Spec.Resource,
			Description: req.Spec.Description,
		}
		apiKey.Metadata = &apikeysv2.ObjectMeta{CreatedAt: keyTime}
		keyIndex++
		keyStoreV2[*apiKey.Id] = apiKey
		err = json.NewEncoder(w).Encode(apiKey)
		require.NoError(t, err)
	}
}

// Handler for: "/iam/v2/users/{id}"
func handleIamUser(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["id"]
		w.Header().Set("Content-Type", "application/json")
		var user iamv2.IamV2User
		switch userId {
		case "u-0", "u-1":
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		case "u-2":
			user = buildIamUser("u-2@confluent.io", "Bono", "u-2", "AUTH_TYPE_LOCAL")
		case "u-11aaa":
			user = buildIamUser("u-11aaa@confluent.io", "11 Aaa", "u-11aaa", "AUTH_TYPE_LOCAL")
		default:
			user = buildIamUser("mhe@confluent.io", "Muwei He", userId, "AUTH_TYPE_LOCAL")
		}
		err := json.NewEncoder(w).Encode(user)
		require.NoError(t, err)
	}
}

// Handler for: "/iam/v2/users"
func handleIamUsers(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			users := []iamv2.IamV2User{
				buildIamUser("bstrauch@confluent.io", "Brian Strauch", "u11", "AUTH_TYPE_LOCAL"),
				buildIamUser("mtodzo@confluent.io", "Miles Todzo", "u-17", "AUTH_TYPE_LOCAL"),
				buildIamUser("u-11aaa@confluent.io", "11 Aaa", "u-11aaa", "AUTH_TYPE_LOCAL"),
				buildIamUser("u-22bbb@confluent.io", "22 Bbb", "u-22bbb", "AUTH_TYPE_LOCAL"),
				buildIamUser("u-33ccc@confluent.io", "33 Ccc", "u-33ccc", "AUTH_TYPE_SSO"),
				buildIamUser("mhe@confluent.io", "Muwei He", "u-44ddd", "AUTH_TYPE_LOCAL"),
			}
			userId := r.URL.Query().Get("id")
			if userId != "" {
				if userId == deactivatedResourceID {
					users = []iamv2.IamV2User{}
				}
			}
			res := iamv2.IamV2UserList{
				Data: users,
			}
			email := r.URL.Query().Get("email")
			if email != "" {
				for _, u := range users {
					if *u.Email == email {
						res = iamv2.IamV2UserList{
							Data: []iamv2.IamV2User{u},
						}
						break
					}
				}
			}
			w.Header().Set("Content-Type", "application/json")
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/service_accounts/{id}"
func handleIamServiceAccount(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodGet:
			switch id {
			case "sa-6789":
				err := writeResourceNotFoundError(w)
				require.NoError(t, err)
			default:
				serviceAccount := iamv2.IamV2ServiceAccount{
					Id:          iamv2.PtrString(serviceAccountResourceID),
					DisplayName: iamv2.PtrString("service_account"),
					Description: iamv2.PtrString("at your service."),
				}
				err := json.NewEncoder(w).Encode(serviceAccount)
				require.NoError(t, err)
			}
		case http.MethodPatch:
			var req iamv2.IamV2ServiceAccount
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamv2.IamV2ServiceAccount{Id: iamv2.PtrString(id), Description: req.Description}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/service_accounts"
func handleIamServiceAccounts(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString(serviceAccountResourceID),
				DisplayName: iamv2.PtrString("service_account"),
				Description: iamv2.PtrString("at your service."),
			}
			err := json.NewEncoder(w).Encode(iamv2.IamV2ServiceAccountList{Data: []iamv2.IamV2ServiceAccount{serviceAccount}})
			require.NoError(t, err)
		case http.MethodPost:
			var req iamv2.IamV2ServiceAccount
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString("sa-55555"),
				DisplayName: req.DisplayName,
				Description: req.Description,
			}
			err = json.NewEncoder(w).Encode(serviceAccount)
			require.NoError(t, err)
		}
	}
}

// Handler for :"/iam/v2/role-bindings"
func handleIamRoleBindings(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		crnPattern := r.URL.Query().Get("crn_pattern")
		principal := r.URL.Query().Get("principal")
		roleName := r.URL.Query().Get("role_name")
		switch r.Method {
		case http.MethodGet:
			if roleName == "InvalidOrgAdmin" || roleName == "InvalidMetricsViewer" {
				err := writeInvalidRoleNameError(w, roleName)
				require.NoError(t, err)
				return
			}
			roleBindings := []mdsv2.IamV2RoleBinding{}
			for _, rolebinding := range roleBindingStore {
				if isRoleBindingMatch(rolebinding, principal, roleName, crnPattern) {
					roleBindings = append(roleBindings, rolebinding)
				}
			}
			res := mdsv2.IamV2RoleBindingList{Data: roleBindings}
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodPost:
			w.WriteHeader(http.StatusCreated)
		}
	}
}

// Handler for: "/iam/v2/identity-provider/{id}"
func handleIamIdentityProvider(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodPatch:
			var req identityproviderv2.IamV2IdentityProvider
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &identityproviderv2.IamV2IdentityProvider{
				Id:          identityproviderv2.PtrString("op-55555"),
				DisplayName: req.DisplayName,
				Description: req.Description,
				Issuer:      identityproviderv2.PtrString("https://company.provider.com"),
				JwksUri:     identityproviderv2.PtrString("https://company.provider.com/oauth2/v1/keys"),
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			switch id {
			case "op-1":
				w.Header().Set("Content-Type", "application/json")
				err := writeResourceNotFoundError(w)
				require.NoError(t, err)
			default:
				w.WriteHeader(http.StatusNoContent)
			}
		case http.MethodGet:
			identityProvider := identityproviderv2.IamV2IdentityProvider{
				Id:          identityproviderv2.PtrString(identityProviderResourceID),
				DisplayName: identityproviderv2.PtrString("identity_provider"),
				Description: identityproviderv2.PtrString("providing identities."),
				Issuer:      identityproviderv2.PtrString("https://company.provider.com"),
				JwksUri:     identityproviderv2.PtrString("https://company.provider.com/oauth2/v1/keys"),
			}
			err := json.NewEncoder(w).Encode(identityProvider)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/identity-providers"
func handleIamIdentityProviders(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			identityProvider := identityproviderv2.IamV2IdentityProvider{
				Id:          identityproviderv2.PtrString(identityProviderResourceID),
				DisplayName: identityproviderv2.PtrString("identity_provider"),
				Description: identityproviderv2.PtrString("providing identities."),
				Issuer:      identityproviderv2.PtrString("https://company.provider.com"),
				JwksUri:     identityproviderv2.PtrString("https://company.provider.com/oauth2/v1/keys"),
			}
			err := json.NewEncoder(w).Encode(identityproviderv2.IamV2IdentityProviderList{Data: []identityproviderv2.IamV2IdentityProvider{identityProvider, identityProvider}})
			require.NoError(t, err)
		case http.MethodPost:
			var req identityproviderv2.IamV2IdentityProvider
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			identityProvider := &identityproviderv2.IamV2IdentityProvider{
				Id:          identityproviderv2.PtrString("op-55555"),
				DisplayName: req.DisplayName,
				Description: req.Description,
				Issuer:      req.Issuer,
				JwksUri:     req.JwksUri,
			}
			err = json.NewEncoder(w).Encode(identityProvider)
			require.NoError(t, err)
		}
	}
}

// Handler for :"/iam/v2/role-bindings/{id}"
func handleIamRoleBinding(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/identity-providers/{provider_id}/identity-pools/{id}"
func handleIamIdentityPool(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		id := mux.Vars(r)["id"]
		switch r.Method {
		case http.MethodPatch:
			var req identityproviderv2.IamV2IdentityPool
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &identityproviderv2.IamV2IdentityPool{
				Id:            identityproviderv2.PtrString("op-55555"),
				DisplayName:   req.DisplayName,
				Description:   req.Description,
				IdentityClaim: req.IdentityClaim,
				Filter:        req.Filter,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			switch id {
			case "pool-1":
				w.Header().Set("Content-Type", "application/json")
				err := writeResourceNotFoundError(w)
				require.NoError(t, err)
			default:
				w.WriteHeader(http.StatusNoContent)
			}
		case http.MethodGet:
			identityPool := identityproviderv2.IamV2IdentityPool{
				Id:            identityproviderv2.PtrString(identityPoolResourceID),
				DisplayName:   identityproviderv2.PtrString("identity_pool"),
				Description:   identityproviderv2.PtrString("pooling identities"),
				IdentityClaim: identityproviderv2.PtrString("sub"),
				Filter:        identityproviderv2.PtrString(`claims.iss="https://company.provider.com"`),
			}
			err := json.NewEncoder(w).Encode(identityPool)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/identity-providers/{provider_id}/identity-pools"
func handleIamIdentityPools(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			identityPool := identityproviderv2.IamV2IdentityPool{
				Id:            identityproviderv2.PtrString(identityPoolResourceID),
				DisplayName:   identityproviderv2.PtrString("identity_pool"),
				Description:   identityproviderv2.PtrString("pooling identities."),
				IdentityClaim: identityproviderv2.PtrString("sub"),
				Filter:        identityproviderv2.PtrString(`claims.iss="https://company.provider.com"`),
			}
			identityPoolTwo := identityproviderv2.IamV2IdentityPool{
				Id:            identityproviderv2.PtrString("pool-55555"),
				DisplayName:   identityproviderv2.PtrString("identity_pool_2"),
				Description:   identityproviderv2.PtrString("pooling identities."),
				IdentityClaim: identityproviderv2.PtrString("sub"),
				Filter:        identityproviderv2.PtrString(`claims.iss="https://company.provider.com"`),
			}
			err := json.NewEncoder(w).Encode(identityproviderv2.IamV2IdentityPoolList{Data: []identityproviderv2.IamV2IdentityPool{identityPool, identityPoolTwo}})
			require.NoError(t, err)
		case http.MethodPost:
			var req identityproviderv2.IamV2IdentityPool
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			identityPool := &identityproviderv2.IamV2IdentityPool{
				Id:            identityproviderv2.PtrString("pool-55555"),
				DisplayName:   req.DisplayName,
				Description:   req.Description,
				IdentityClaim: req.IdentityClaim,
				Filter:        req.Filter,
			}
			err = json.NewEncoder(w).Encode(identityPool)
			require.NoError(t, err)
		}
	}
}

// Handler for "iam/v2/invitations"
func handleIamInvitations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.Method {
		case http.MethodGet:
			invitationList := &iamv2.IamV2InvitationList{Data: []iamv2.IamV2Invitation{
				buildIamInvitation("1", "u-11aaa@confluent.io", "u-11aaa", "VERIFIED"),
				buildIamInvitation("2", "u-22bbb@confluent.io", "u-22bbb", "SENT"),
			}}
			err := json.NewEncoder(w).Encode(invitationList)
			require.NoError(t, err)
		case http.MethodPost:
			req := new(iamv2.IamV2Invitation)
			err := json.NewDecoder(r.Body).Decode(req)
			require.NoError(t, err)
			if strings.Contains(req.GetEmail(), "user@exists.com") {
				err = writeUserConflictError(w)
				require.NoError(t, err)
			} else {
				invitation := buildIamInvitation("1", "miles@confluent.io", "user1", "SENT")
				err = json.NewEncoder(w).Encode(invitation)
				require.NoError(t, err)
			}
		}
	}
}
