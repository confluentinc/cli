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
	certificateauthorityv2 "github.com/confluentinc/ccloud-sdk-go-v2/certificate-authority/v2"
	iamipfilteringv2 "github.com/confluentinc/ccloud-sdk-go-v2/iam-ip-filtering/v2"
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
		buildRoleBinding("rb-11111", "u-11111", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("rb-12345", "sa-12345", "OrganizationAdmin",
			"crn://confluent.cloud/organization=abc-123"),
		buildRoleBinding("rb-123ab", "group-abc", "OrganizationAdmin",
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
			setPageToken(apiKeysList, &apiKeysList.Metadata, r.URL)
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
	if id == "tableflow" {
		return "Tableflow"
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
			setPageToken(&res, &res.Metadata, r.URL)
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
			serviceAccounts := &iamv2.IamV2ServiceAccountList{
				Data: []iamv2.IamV2ServiceAccount{
					{
						Id:          iamv2.PtrString(serviceAccountResourceId),
						DisplayName: iamv2.PtrString("service-account"),
						Description: iamv2.PtrString("at your service."),
					},
				},
			}
			setPageToken(serviceAccounts, &serviceAccounts.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(serviceAccounts)
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
			setPageToken(&res, &res.Metadata, r.URL)
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
			if id == "op-67890" {
				res.IdentityClaim = req.IdentityClaim
				res.DisplayName = identityproviderv2.PtrString("okta-with-identity-claim")
				res.Description = identityproviderv2.PtrString("providing identities with identity claim.")
				res.Issuer = identityproviderv2.PtrString("https://company.new-provider.com")
				res.JwksUri = identityproviderv2.PtrString("https://company.new-provider.com/oauth2/v1/keys")
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			if id == identityProviderId {
				identityProvider := buildIamProvider(id, "identity-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys", "")
				err := json.NewEncoder(w).Encode(identityProvider)
				require.NoError(t, err)
			} else if id == "op-67890" {
				identityProviderIdentityClaim := buildIamProvider(id, "okta-with-identity-claim", "new description.", "https://company.new-provider.com", "https://company.new-provider.com/oauth2/v1/keys", "claims.sub")
				err := json.NewEncoder(w).Encode(identityProviderIdentityClaim)
				require.NoError(t, err)
			}
		}
	}
}

// Handler for: "/iam/v2/identity-providers"
func handleIamIdentityProviders(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			identityProvider := buildIamProvider(identityProviderId, "identity-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys", "")
			anotherIdentityProvider := buildIamProvider("op-abc", "another-provider", "providing identities.", "https://company.provider.com", "https://company.provider.com/oauth2/v1/keys", "")
			identityProviderIdentityClaim := buildIamProvider("op-67890", "okta-with-identity-claim", "new description.", "https://company.new-provider.com", "https://company.new-provider.com/oauth2/v1/keys", "claims.sub")
			identityProviderList := &identityproviderv2.IamV2IdentityProviderList{Data: []identityproviderv2.IamV2IdentityProvider{identityProvider, anotherIdentityProvider, identityProviderIdentityClaim}}
			setPageToken(identityProviderList, &identityProviderList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(identityProviderList)
			require.NoError(t, err)
		case http.MethodPost:
			var req identityproviderv2.IamV2IdentityProvider
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			identityProvider := &identityproviderv2.IamV2IdentityProvider{
				DisplayName: req.DisplayName,
				Description: req.Description,
				Issuer:      req.Issuer,
				JwksUri:     req.JwksUri,
			}
			if *req.DisplayName == "okta" {
				identityProvider.Id = identityproviderv2.PtrString("op-55555")
			} else if *req.DisplayName == "okta-with-identity-claim" {
				identityProvider.Id = identityproviderv2.PtrString("op-67890")
				identityProvider.IdentityClaim = req.IdentityClaim
				identityProvider.Description = identityproviderv2.PtrString("new description.")
				identityProvider.Issuer = identityproviderv2.PtrString("https://company.new-provider.com")
				identityProvider.JwksUri = identityproviderv2.PtrString("https://company.new-provider.com/oauth2/v1/keys")
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
			identityPoolList := &identityproviderv2.IamV2IdentityPoolList{Data: []identityproviderv2.IamV2IdentityPool{identityPool, anotherIdentityPool}}
			setPageToken(identityPoolList, &identityPoolList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(identityPoolList)
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

// Handler for: "/iam/v2/certificate-authorities/{id}"
func handleIamCertificateAuthority(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != "op-12345" && id != "op-54321" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodGet:
			certificateAuthority := buildIamCertificateAuthority(id, "my-ca", "my certificate authority", "certificate.pem", "", "")
			err := json.NewEncoder(w).Encode(certificateAuthority)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodPut:
			var req certificateauthorityv2.IamV2UpdateCertRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			certificateAuthority := buildIamCertificateAuthority(id, req.GetDisplayName(), req.GetDescription(), req.GetCertificateChainFilename(), req.GetCrlUrl(), req.GetCrlChain())
			err = json.NewEncoder(w).Encode(certificateAuthority)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/certificate-authorities"
func handleIamCertificateAuthorities(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			certificateAuthorityList := &certificateauthorityv2.IamV2CertificateAuthorityList{Data: []certificateauthorityv2.IamV2CertificateAuthority{
				buildIamCertificateAuthority("op-12345", "my-ca", "my certificate authority", "certificate.pem", "", ""),
				buildIamCertificateAuthority("op-54321", "my-ca-2", "my other certificate authority", "certificate-2.pem", "", "DEF456"),
				buildIamCertificateAuthority("op-67890", "my-ca-3", "my other certificate authority", "certificate-3.pem", "example.url", ""),
			}}
			setPageToken(certificateAuthorityList, &certificateAuthorityList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(certificateAuthorityList)
			require.NoError(t, err)
		case http.MethodPost:
			var req certificateauthorityv2.IamV2CreateCertRequest
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			certificateAuthority := buildIamCertificateAuthority("op-12345", req.GetDisplayName(), req.GetDescription(), req.GetCertificateChainFilename(), req.GetCrlUrl(), req.GetCrlChain())
			err = json.NewEncoder(w).Encode(certificateAuthority)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/certificate-authorities/{provider_id}/identity-pools/{id}"
func handleIamCertificatePool(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != identityPoolId && id != "pool-55555" {
			err := writeResourceNotFoundError(w)
			require.NoError(t, err)
			return
		}
		switch r.Method {
		case http.MethodPut:
			var req certificateauthorityv2.IamV2CertificateIdentityPool
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &certificateauthorityv2.IamV2CertificateIdentityPool{
				Id:                 req.Id,
				DisplayName:        req.DisplayName,
				Description:        req.Description,
				ExternalIdentifier: req.ExternalIdentifier,
				Filter:             req.Filter,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodDelete:
			w.WriteHeader(http.StatusNoContent)
		case http.MethodGet:
			certificatePool := buildIamCertificatePool(id, "certificate-pool", "certificate pool identities", "external-entity", "true")
			err := json.NewEncoder(w).Encode(certificatePool)
			require.NoError(t, err)
		}
	}
}

// Handler for: "/iam/v2/certificate-authorities/{provider_id}/identity-pools"
func handleIamCertificatePools(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			certificatePool := buildIamCertificatePool(identityPoolId, "certificate-pool", "certificate pool identities", "external-entity", "true")
			anotherCertificatePool := buildIamCertificatePool("pool-abc", "another-pool", "another description", "sub", "true")
			certificatePoolList := &certificateauthorityv2.IamV2CertificateIdentityPoolList{Data: []certificateauthorityv2.IamV2CertificateIdentityPool{certificatePool, anotherCertificatePool}}
			setPageToken(certificatePoolList, &certificatePoolList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(certificatePoolList)
			require.NoError(t, err)
		case http.MethodPost:
			var req certificateauthorityv2.IamV2CertificateIdentityPool
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			certificatePool := &certificateauthorityv2.IamV2CertificateIdentityPool{
				Id:                 certificateauthorityv2.PtrString("pool-55555"),
				DisplayName:        req.DisplayName,
				Description:        req.Description,
				ExternalIdentifier: req.ExternalIdentifier,
				Filter:             req.Filter,
			}
			err = json.NewEncoder(w).Encode(certificatePool)
			require.NoError(t, err)
		}
	}
}

// Handler for: "iam/v2/ip-filters"
func handleIamIpFilters(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			ipFilter := buildIamIpFilter(ipFilterId, "demo-ip-filter", "multiple", []string{"ipg-12345", "ipg-abcde"}, "crn://confluent.cloud/organization=org123", []string{"MANAGEMENT"})
			ipFilterList := &iamipfilteringv2.IamV2IpFilterList{Data: []iamipfilteringv2.IamV2IpFilter{ipFilter}}
			setPageToken(ipFilterList, &ipFilterList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(ipFilterList)
			require.NoError(t, err)
		case http.MethodPost:
			var req iamipfilteringv2.IamV2IpFilter
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			opGroups := req.OperationGroups
			if *req.ResourceGroup == "management" {
				opGroups = &[]string{"MANAGEMENT"}
				req.ResourceGroup = iamipfilteringv2.PtrString("multiple")
			}
			ipFilter := &iamipfilteringv2.IamV2IpFilter{
				Id:              iamipfilteringv2.PtrString(ipFilterId),
				FilterName:      req.FilterName,
				ResourceGroup:   req.ResourceGroup,
				IpGroups:        req.IpGroups,
				ResourceScope:   iamipfilteringv2.PtrString("crn://confluent.cloud/organization=org123"),
				OperationGroups: opGroups,
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
			var req iamipfilteringv2.IamV2IpFilter
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamipfilteringv2.IamV2IpFilter{
				Id:              req.Id,
				FilterName:      req.FilterName,
				ResourceGroup:   req.ResourceGroup,
				IpGroups:        req.IpGroups,
				ResourceScope:   req.ResourceScope,
				OperationGroups: req.OperationGroups,
			}
			err = json.NewEncoder(w).Encode(res)
			require.NoError(t, err)
		case http.MethodGet:
			var ipFilter iamipfilteringv2.IamV2IpFilter
			segments := strings.Split(r.URL.String(), "/")
			var filterId string
			if len(segments) > 0 {
				filterId = segments[len(segments)-1] // "ipf-34dq3"
			}
			var operationGroups []string
			if filterId == "ipf-34dq4" {
				operationGroups = []string{"MANAGEMENT", "SCHEMA"}
			} else if filterId == "ipf-34dq6" {
				operationGroups = []string{"MANAGEMENT", "SCHEMA", "FLINK"}
			} else if filterId == "ipf-34dq7" {
				operationGroups = []string{"KAFKA_MANAGEMENT", "KAFKA_DATA", "KAFKA_DISCOVERY"}
			} else {
				operationGroups = []string{"MANAGEMENT"}
			}
			ipFilter = buildIamIpFilter(ipFilterId, "demo-ip-filter", "multiple", []string{"ipg-12345", "ipg-abcde"}, "crn://confluent.cloud/organization=org123", operationGroups)
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
			ipGroupList := &iamipfilteringv2.IamV2IpGroupList{Data: []iamipfilteringv2.IamV2IpGroup{ipGroup}}
			setPageToken(ipGroupList, &ipGroupList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(ipGroupList)
			require.NoError(t, err)
		case http.MethodPost:
			var req iamipfilteringv2.IamV2IpGroup
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			ipGroup := &iamipfilteringv2.IamV2IpGroup{
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
			var req iamipfilteringv2.IamV2IpGroup
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			res := &iamipfilteringv2.IamV2IpGroup{
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
			setPageToken(invitationList, &invitationList.Metadata, r.URL)
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
		groupMapping := buildIamGroupMapping("group-12345", "my-group-mapping", "new description", `"engineering" in claims.group || "marketing" in claims.group`)
		switch r.Method {
		case http.MethodGet:
			anotherMapping := buildIamGroupMapping(groupMappingId, "another-group-mapping", "another description", "true")
			groupMappingsList := &ssov2.IamV2SsoGroupMappingList{Data: []ssov2.IamV2SsoGroupMapping{groupMapping, anotherMapping}}
			setPageToken(groupMappingsList, &groupMappingsList.Metadata, r.URL)
			err := json.NewEncoder(w).Encode(groupMappingsList)
			require.NoError(t, err)
		case http.MethodPost:
			var req ssov2.IamV2SsoGroupMapping
			err := json.NewDecoder(r.Body).Decode(&req)
			require.NoError(t, err)
			if req.GetDisplayName() == "group-mapping-rate-limit" {
				w.WriteHeader(http.StatusPaymentRequired)
				err = writeErrorJson(w, "Group mapping reached limit 12")
				require.NoError(t, err)
			} else {
				err = json.NewEncoder(w).Encode(&groupMapping)
				require.NoError(t, err)
			}
		}
	}
}

// Handler for "iam/v2/sso/group-mappings/{id}"
func handleIamGroupMapping(t *testing.T) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := mux.Vars(r)["id"]
		if id != groupMappingId && id != "group-def" {
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

func buildIamCertificateAuthority(id, name, description, certificateChainFilename, crlUrl, crlChain string) certificateauthorityv2.IamV2CertificateAuthority {
	expDate, _ := time.Parse(time.RFC3339, "2017-07-21T17:32:28Z")

	crlSource := ""
	if crlUrl != "" {
		crlSource = "URL"
	}
	if crlChain != "" {
		crlUrl = "" // prefer chain over url
		crlSource = "LOCAL"
	}

	var crlUpdatedAt *time.Time
	if crlUrl != "" || crlChain != "" {
		updatedAt, _ := time.Parse(time.RFC3339, "2024-07-21T17:32:28Z")
		crlUpdatedAt = &updatedAt
	}

	return certificateauthorityv2.IamV2CertificateAuthority{
		Id:                       certificateauthorityv2.PtrString(id),
		DisplayName:              certificateauthorityv2.PtrString(name),
		Description:              certificateauthorityv2.PtrString(description),
		Fingerprints:             &[]string{"B1BC968BD4f49D622AA89A81F2150152A41D829C"},
		ExpirationDates:          &[]time.Time{expDate},
		SerialNumbers:            &[]string{"219C542DE8f6EC7177FA4EE8C3705797"},
		CertificateChainFilename: certificateauthorityv2.PtrString(certificateChainFilename),
		CrlSource:                certificateauthorityv2.PtrString(crlSource),
		CrlUrl:                   certificateauthorityv2.PtrString(crlUrl),
		CrlUpdatedAt:             crlUpdatedAt,
	}
}

func buildIamCertificatePool(id, name, description, externalIdentifier, filter string) certificateauthorityv2.IamV2CertificateIdentityPool {
	return certificateauthorityv2.IamV2CertificateIdentityPool{
		Id:                 iamv2.PtrString(id),
		DisplayName:        iamv2.PtrString(name),
		Description:        iamv2.PtrString(description),
		ExternalIdentifier: iamv2.PtrString(externalIdentifier),
		Filter:             certificateauthorityv2.PtrString(filter),
	}
}

func buildIamProvider(id, name, description, issuer, jwksUri, identityClaim string) identityproviderv2.IamV2IdentityProvider {
	return identityproviderv2.IamV2IdentityProvider{
		Id:            iamv2.PtrString(id),
		DisplayName:   iamv2.PtrString(name),
		Description:   iamv2.PtrString(description),
		IdentityClaim: iamv2.PtrString(identityClaim),
		Issuer:        iamv2.PtrString(issuer),
		JwksUri:       iamv2.PtrString(jwksUri),
	}
}

func buildIamIpFilter(id string, name string, resourceGroup string, ipGroupIds []string, resourceScope string, operationGroups []string) iamipfilteringv2.IamV2IpFilter {
	// Convert the IP group IDs into IP group objects
	IpGroupIdObjects := make([]iamipfilteringv2.GlobalObjectReference, len(ipGroupIds))
	for i, ipGroupId := range ipGroupIds {
		// The empty string fields will get filled in automatically by the cc-policy-service
		IpGroupIdObjects[i] = iamipfilteringv2.GlobalObjectReference{Id: ipGroupId}
	}

	return iamipfilteringv2.IamV2IpFilter{
		Id:              iamipfilteringv2.PtrString(id),
		FilterName:      iamipfilteringv2.PtrString(name),
		ResourceGroup:   iamipfilteringv2.PtrString(resourceGroup),
		IpGroups:        &IpGroupIdObjects,
		OperationGroups: &operationGroups,
		ResourceScope:   &resourceScope,
	}
}

func buildIamIpGroup(id string, name string, cidrBlocks []string) iamipfilteringv2.IamV2IpGroup {
	return iamipfilteringv2.IamV2IpGroup{
		Id:         iamipfilteringv2.PtrString(id),
		GroupName:  iamipfilteringv2.PtrString(name),
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
