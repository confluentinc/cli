package testserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	apikeysv2 "github.com/confluentinc/ccloud-sdk-go-v2/apikeys/v2"
	iamv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam/v2"
	identityproviderv2 "github.com/confluentinc/ccloud-sdk-go-v2/identity-provider/v2"
	mdsv2 "github.com/confluentinc/ccloud-sdk-go-v2/mds/v2"
	ssov2 "github.com/confluentinc/ccloud-sdk-go-v2/sso/v2"
)

var (
	keyStoreV2       = map[string]*apikeysv2.IamV2ApiKey{}
	keyTime          = apikeysv2.PtrTime(time.Date(1999, time.February, 24, 0, 0, 0, 0, time.UTC))
	roleBindingStore = []mdsv2.IamV2RoleBinding{
		buildRoleBinding("rb-00000", identityPoolId, "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123/identity-provider="+identityProviderId),
		buildRoleBinding("rb-11aaa", "u-11aaa", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("rb-12345", "sa-12345", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("rb-111aa", "u-11aaa", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("rb-22bbb", "u-22bbb", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("rb-222bb", "u-22bbb", "EnvironmentAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=env-596"),
		buildRoleBinding("rb-33ccc", "u-33ccc", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("rb-44ddd", "u-44ddd", "CloudClusterAdmin",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa"),
		buildRoleBinding("rb-55eee", "u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/group=readers"),
		buildRoleBinding("rb-555ee", "u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/topic=clicks-*"),
		buildRoleBinding("rb-5555e", "u-55eee", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa/kafka=lkc-1111aaa/topic=payroll"),
		buildRoleBinding("rb-66fff", "u-66fff", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/cloud-cluster=lkc-1111aaa/ksql=ksql-cluster-name-2222bbb"),
		buildRoleBinding("rb-77ggg", "u-77ggg", "ResourceOwner",
			"crn://confluent.cloud/organization=abc-123/environment=env-596/schema-registry=lsrc-3333ccc/subject=clicks"),
	}
)

func init() {
	fillKeyStoreV2()
}

// Handler for: "/iam/v2/api-keys/{id}"
func handleIamApiKey(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
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
		if apiKey, ok := keyStoreV2[keyStr]; ok {
			err := json.NewEncoder(w).Encode(apiKey)
			require.NoError(t, err)
		} else {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
		}
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
		req := &apikeysv2.IamV2ApiKey{}
		err := json.NewDecoder(r.Body).Decode(req)
		require.NoError(t, err)

		if req.Spec.Owner.Id == "sa-123456" {
			err = writeServiceAccountInvalidError(w)
			require.NoError(t, err)
			return
		}

		apiKey := req

		switch req.Spec.Resource.GetKind() {
		case "Region":
			apiKey = &apikeysv2.IamV2ApiKey{
				Id:         apikeysv2.PtrString("FLINKREGIONAPIKEY"),
				ApiVersion: apikeysv2.PtrString("fcpm/v2"),
				Spec:       &apikeysv2.IamV2ApiKeySpec{Secret: apikeysv2.PtrString("FLINKREGIONAPISECRET")},
			}
		default:
			apiKey.Id = apikeysv2.PtrString(fmt.Sprintf("MYKEY%d", keyIndex))
			apiKey.Spec = &apikeysv2.IamV2ApiKeySpec{
				Owner:  req.Spec.Owner,
				Secret: apikeysv2.PtrString(fmt.Sprintf("MYSECRET%d", keyIndex)),
				Resource: &apikeysv2.ObjectReference{
					Id:         req.Spec.Resource.GetId(),
					ApiVersion: apikeysv2.PtrString("cmk/v2"),
					Kind:       apikeysv2.PtrString(getKind(req.Spec.Resource.GetId())),
				},
				Description: req.Spec.Description,
			}
			apiKey.Metadata = &apikeysv2.ObjectMeta{CreatedAt: keyTime}
			keyIndex++
			keyStoreV2[apiKey.GetId()] = apiKey
		}

		err = json.NewEncoder(w).Encode(apiKey)
		require.NoError(t, err)
	}
}

func getKind(id string) string {
	if id == "cloud" {
		return "Cloud"
	}

	x := strings.SplitN(id, "-", 2)
	if len(x) != 2 {
		return ""
	}

	prefixToKind := map[string]string{
		"lkc":    "Cluster",
		"lksqlc": "ksqlDB",
		"lsrc":   "SchemaRegistry",
	}

	kind, ok := prefixToKind[x[0]]
	if !ok {
		return ""
	}

	return kind
}

// Handler for: "/iam/v2/users/{id}"
func handleIamUser(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		userId := vars["id"]
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
				if userId == deactivatedUserResourceId {
					users = []iamv2.IamV2User{}
				}
			}
			res := iamv2.IamV2UserList{Data: users}
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
			err := json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/service_accounts/{id}"
func handleIamServiceAccount(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != serviceAccountResourceId && id != "sa-54321" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString(id),
				DisplayName: iamv2.PtrString("service-account"),
				Description: iamv2.PtrString("at your service."),
			}
			err := json.NewEncoder(w).Encode(serviceAccount)
			require.NoError(t, err)
		case http.MethodPatch:
			var req iamv2.IamV2ServiceAccount
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamv2.IamV2ServiceAccount{Id: iamv2.PtrString(req.GetId()), Description: req.Description}
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
		switch r.Method {
		case http.MethodGet:
			serviceAccount := iamv2.IamV2ServiceAccount{
				Id:          iamv2.PtrString(serviceAccountResourceId),
				DisplayName: iamv2.PtrString("service-account"),
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
		id := mux.Vars(r)["id"]
		if id != identityProviderId && id != "op-67890" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var req identityproviderv2.IamV2IdentityProvider
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &identityproviderv2.IamV2IdentityProvider{
				Id:          req.Id,
				DisplayName: req.DisplayName,
				Description: req.Description,
				Issuer:      identityproviderv2.PtrString("https://company.provider.com"),
				JwksUri:     identityproviderv2.PtrString("https://company.provider.com/oauth2/v1/keys"),
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			identityProvider := buildIamProvider(id, "identity-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys")
			err := json.NewEncoder(w).Encode(identityProvider)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/identity-providers"
func handleIamIdentityProviders(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			identityProvider := buildIamProvider(identityProviderId, "identity-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys")
			anotherIdentityProvider := buildIamProvider("op-abc", "another-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys")
			err := json.NewEncoder(w).Encode(identityproviderv2.IamV2IdentityProviderList{Data: []identityproviderv2.IamV2IdentityProvider{identityProvider, anotherIdentityProvider}})
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
func handleIamRoleBinding(_ *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/identity-providers/{provider_id}/identity-pools/{id}"
func handleIamIdentityPool(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != identityPoolId && id != "pool-55555" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var req identityproviderv2.IamV2IdentityPool
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &identityproviderv2.IamV2IdentityPool{
				Id:            req.Id,
				DisplayName:   req.DisplayName,
				Description:   req.Description,
				IdentityClaim: req.IdentityClaim,
				Filter:        req.Filter,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			identityPool := buildIamPool(id, "identity-pool", "pooling identities", "sub", `claims.iss="https://company.provider.com"`)
			err := json.NewEncoder(w).Encode(identityPool)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/identity-providers/{provider_id}/identity-pools"
func handleIamIdentityPools(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			identityPool := buildIamPool(identityPoolId, "identity-pool", "pooling identities", "sub", `claims.iss="https://company.provider.com"`)
			anotherIdentityPool := buildIamPool("pool-abc", "another-pool", "another description", "sub", "true")
			err := json.NewEncoder(w).Encode(identityproviderv2.IamV2IdentityPoolList{Data: []identityproviderv2.IamV2IdentityPool{identityPool, anotherIdentityPool}})
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

// Handler for: "iam/v2/ip-filters"
func handleIamIpFilters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ipFilter := buildIamIpFilter(ipFilterId, "demo-ip-filter", "management", []string{"ipg-12345", "ipg-abcde"})
			err := json.NewEncoder(w).Encode(iamv2.IamV2IpFilterList{Data: []iamv2.IamV2IpFilter{ipFilter}})
			require.NoError(t, err)
		case http.MethodPost:
			var req iamv2.IamV2IpFilter
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)

			ipFilter := &iamv2.IamV2IpFilter{
				Id:            iamv2.PtrString(ipFilterId),
				FilterName:    req.FilterName,
				ResourceGroup: req.ResourceGroup,
				IpGroups:      req.IpGroups,
			}
			err = json.NewEncoder(w).Encode(ipFilter)
			require.NoError(t, err)
		}
	}
}

// Handler for: "iam/iv2/ip-filters/{id}"
func handleIamIpFilter(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			var req iamv2.IamV2IpFilter
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamv2.IamV2IpFilter{
				Id:            req.Id,
				FilterName:    req.FilterName,
				ResourceGroup: req.ResourceGroup,
				IpGroups:      req.IpGroups,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet:
			ipFilter := buildIamIpFilter(ipFilterId, "demo-ip-filter", "management", []string{"ipg-12345", "ipg-abcde"})
			err := json.NewEncoder(w).Encode(ipFilter)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for: "/iam/v2/ip-groups"
func handleIamIpGroups(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ipGroup := buildIamIpGroup(ipGroupId, "demo-ip-group", []string{"168.150.200.0/24", "147.150.200.0/24"})
			err := json.NewEncoder(w).Encode(iamv2.IamV2IpGroupList{Data: []iamv2.IamV2IpGroup{ipGroup}})
			require.NoError(t, err)
		case http.MethodPost:
			var req iamv2.IamV2IpGroup
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			ipGroup := &iamv2.IamV2IpGroup{
				Id:         iamv2.PtrString(ipGroupId),
				GroupName:  req.GroupName,
				CidrBlocks: req.CidrBlocks,
			}
			err = json.NewEncoder(w).Encode(ipGroup)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/ip-groups/{id}"
func handleIamIpGroup(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPatch:
			var req iamv2.IamV2IpGroup
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamv2.IamV2IpGroup{
				Id:         req.Id,
				GroupName:  req.GroupName,
				CidrBlocks: req.CidrBlocks,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet:
			ipGroup := buildIamIpGroup(ipGroupId, "demo-ip-group", []string{"168.150.200.0/24", "147.150.200.0/24"})
			err := json.NewEncoder(w).Encode(ipGroup)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		}
	}
}

// Handler for "iam/v2/invitations"
func handleIamInvitations(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			invitationList := &iamv2.IamV2InvitationList{Data: []iamv2.IamV2Invitation{
				buildIamInvitation("i-11111", "u-11aaa@confluent.io", "u-11aaa", "VERIFIED"),
				buildIamInvitation("i-22222", "u-22bbb@confluent.io", "u-22bbb", "SENT"),
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

// Handler for "iam/v2/sso/group-mappings"
func handleIamGroupMappings(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		groupMapping := buildIamGroupMapping("pool-12345", "my-group-mapping", "new description", `"engineering" in claims.group || "marketing" in claims.group`)
		switch r.Method {
		case http.MethodGet:
			anotherMapping := buildIamGroupMapping(groupMappingId, "another-group-mapping", "another description", "true")
			err := json.NewEncoder(w).Encode(ssov2.IamV2SsoGroupMappingList{Data: []ssov2.IamV2SsoGroupMapping{groupMapping, anotherMapping}})
			require.NoError(t, err)
		case http.MethodPost:
			var req ssov2.IamV2SsoGroupMapping
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			err = json.NewEncoder(w).Encode(&groupMapping)
			require.NoError(t, err)
		}
	}
}

// Handler for "iam/v2/sso/group-mappings/{id}"
func handleIamGroupMapping(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != groupMappingId && id != "pool-def" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodPatch:
			var req ssov2.IamV2SsoGroupMapping
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := buildIamGroupMapping(req.GetId(), req.GetDisplayName(), req.GetDescription(), req.GetFilter())
			err = json.NewEncoder(w).Encode(&res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			groupMapping := buildIamGroupMapping(id, "another-group-mapping", "another description", "true")
			err := json.NewEncoder(w).Encode(groupMapping)
			require.NoError(t, err)
		}
	}
}

func buildIamUser(email, fullName, id, authType string) iamv2.IamV2User {
	return iamv2.IamV2User{
		Email:    iamv2.PtrString(email),
		FullName: iamv2.PtrString(fullName),
		Id:       iamv2.PtrString(id),
		AuthType: iamv2.PtrString(authType),
	}
}

func buildIamInvitation(id, email, userId, status string) iamv2.IamV2Invitation {
	return iamv2.IamV2Invitation{
		Id:     iamv2.PtrString(id),
		Email:  iamv2.PtrString(email),
		User:   &iamv2.GlobalObjectReference{Id: userId},
		Status: iamv2.PtrString(status),
	}
}

func buildIamGroupMapping(id, name, description, filter string) ssov2.IamV2SsoGroupMapping {
	return ssov2.IamV2SsoGroupMapping{
		Description: ssov2.PtrString(description),
		DisplayName: ssov2.PtrString(name),
		Id:          ssov2.PtrString(id),
		Filter:      ssov2.PtrString(filter),
		Principal:   ssov2.PtrString(id),
		State:       ssov2.PtrString("ENABLED"),
	}
}

func buildIamPool(id, name, description, identityClaim, filter string) identityproviderv2.IamV2IdentityPool {
	return identityproviderv2.IamV2IdentityPool{
		Id:            iamv2.PtrString(id),
		DisplayName:   iamv2.PtrString(name),
		Description:   iamv2.PtrString(description),
		IdentityClaim: iamv2.PtrString(identityClaim),
		Filter:        ssov2.PtrString(filter),
	}
}

func buildIamProvider(id, name, description, issuer, jwksUri string) identityproviderv2.IamV2IdentityProvider {
	return identityproviderv2.IamV2IdentityProvider{
		Id:          iamv2.PtrString(id),
		DisplayName: iamv2.PtrString(name),
		Description: iamv2.PtrString(description),
		Issuer:      iamv2.PtrString(issuer),
		JwksUri:     iamv2.PtrString(jwksUri),
	}
}

func buildIamIpFilter(id string, name string, resourceGroup string, ipGroupIds []string) iamv2.IamV2IpFilter {
	// Convert the IP group IDs into IP group objects
	IpGroupIdObjects := make([]iamv2.GlobalObjectReference, len(ipGroupIds))
	for i, ipGroupId := range ipGroupIds {
		// The empty string fields will get filled in automatically by the cc-policy-service
		IpGroupIdObjects[i] = iamv2.GlobalObjectReference{Id: ipGroupId}
	}

	return iamv2.IamV2IpFilter{
		Id:            iamv2.PtrString(id),
		FilterName:    iamv2.PtrString(name),
		ResourceGroup: iamv2.PtrString(resourceGroup),
		IpGroups:      &IpGroupIdObjects,
	}
}

func buildIamIpGroup(id string, name string, cidrBlocks []string) iamv2.IamV2IpGroup {
	return iamv2.IamV2IpGroup{
		Id:         iamv2.PtrString(id),
		GroupName:  iamv2.PtrString(name),
		CidrBlocks: &cidrBlocks,
	}
}

func buildRoleBinding(id, user, roleName, crn string) mdsv2.IamV2RoleBinding {
	return mdsv2.IamV2RoleBinding{
		Id:         mdsv2.PtrString(id),
		Principal:  mdsv2.PtrString("User:" + user),
		RoleName:   mdsv2.PtrString(roleName),
		CrnPattern: mdsv2.PtrString(crn),
	}
}

func isRoleBindingMatch(rolebinding mdsv2.IamV2RoleBinding, principal, roleName, crnPattern string) bool {
	if !strings.Contains(*rolebinding.CrnPattern, strings.TrimSuffix(crnPattern, "/*")) {
		return false
	}
	if principal != "" && principal != *rolebinding.Principal {
		return false
	}
	if roleName != "" && roleName != *rolebinding.RoleName {
		return false
	}
	return true
}
