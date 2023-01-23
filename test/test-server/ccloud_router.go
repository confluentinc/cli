package testserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

// cloud urls
const (
	sessions            = "/api/sessions"
	me                  = "/api/me"
	loginRealm          = "/api/login/realm"
	account             = "/api/accounts/{id}"
	accounts            = "/api/accounts"
	envMetadata         = "/api/env_metadata"
	serviceAccounts     = "/api/service_accounts"
	serviceAccount      = "/api/service_accounts/{id}"
	schemaRegistries    = "/api/schema_registries"
	schemaRegistry      = "/api/schema_registries/{id}"
	priceTable          = "/api/organizations/{id}/price_table"
	paymentInfo         = "/api/organizations/{id}/payment_info"
	promoCodeClaims     = "/api/organizations/{id}/promo_code_claims"
	users               = "/api/users"
	v2alphaAuthenticate = "/api/metadata/security/v2alpha1/authenticate"
	signup              = "/api/signup"
	launchDarklyProxy   = "/ldapi/sdk/eval/{env}/users/{user:[a-zA-Z0-9=\\-\\/]+}"
	externalIdentities  = "/api/external_identities"
	freeTrialInfo       = "/api/growth/v1/free-trial-info"
)

type CloudRouter struct {
	*mux.Router
	srApiUrl   string
	kafkaRPUrl string
}

// New CloudRouter with all cloud handlers
func NewCloudRouter(t *testing.T, isAuditLogEnabled bool) *CloudRouter {
	c := &CloudRouter{Router: mux.NewRouter()}
	c.buildCcloudRouter(t, isAuditLogEnabled)
	return c
}

// Add handlers for cloud endpoints
func (c *CloudRouter) buildCcloudRouter(t *testing.T, isAuditLogEnabled bool) {
	c.HandleFunc(sessions, handleLogin(t))
	c.HandleFunc(me, c.HandleMe(t, isAuditLogEnabled))
	c.HandleFunc(loginRealm, handleLoginRealm(t))
	c.HandleFunc(signup, c.HandleSignup(t))
	c.HandleFunc(envMetadata, c.HandleEnvMetadata(t))
	c.HandleFunc(launchDarklyProxy, c.HandleLaunchDarkly(t))
	c.HandleFunc(externalIdentities, handleExternalIdentities(t))
	c.addSchemaRegistryRoutes(t)
	c.addEnvironmentRoutes(t)
	c.addOrgRoutes(t)
	c.addUserRoutes(t)
	c.addV2AlphaRoutes(t)
	c.addServiceAccountRoutes(t)
	c.addGrowthRoutes(t)
}

func (c CloudRouter) addV2AlphaRoutes(t *testing.T) {
	c.HandleFunc(v2alphaAuthenticate, c.HandleV2Authenticate(t))
	c.addRoutesAndReplies(t, v2Base, v2RoutesAndReplies)
}

func (c CloudRouter) addRoutesAndReplies(t *testing.T, base string, routesAndReplies map[string]string) {
	jsonRolesMap := rolesListToJsonMap(rbacPublicRoles())
	addRoles(base, routesAndReplies, jsonRolesMap)

	for route, reply := range routesAndReplies {
		s := reply
		c.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			require.NoError(t, err)
		})
	}

	c.HandleFunc(base, c.HandleAllRolesRoute(t))
}

func (c *CloudRouter) addSchemaRegistryRoutes(t *testing.T) {
	c.HandleFunc(schemaRegistries, c.HandleSchemaRegistries(t))
	c.HandleFunc(schemaRegistry, c.HandleSchemaRegistry(t))
}

func (c *CloudRouter) addUserRoutes(t *testing.T) {
	c.HandleFunc(users, c.HandleUsers(t))
}

func (c *CloudRouter) addOrgRoutes(t *testing.T) {
	c.HandleFunc(priceTable, c.HandlePriceTable(t))
	c.HandleFunc(paymentInfo, c.HandlePaymentInfo(t))
	c.HandleFunc(promoCodeClaims, c.HandlePromoCodeClaims(t))
}

func (c *CloudRouter) addEnvironmentRoutes(t *testing.T) {
	c.HandleFunc(accounts, c.HandleEnvironments(t))
	c.HandleFunc(account, c.HandleEnvironment(t))
}

func (c *CloudRouter) addServiceAccountRoutes(t *testing.T) {
	c.HandleFunc(serviceAccounts, c.HandleServiceAccounts(t))
	c.HandleFunc(serviceAccount, c.HandleServiceAccount(t))
}

func (c *CloudRouter) addGrowthRoutes(t *testing.T) {
	c.HandleFunc(freeTrialInfo, c.HandleFreeTrialInfo(t))
}
