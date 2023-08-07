package testserver

import (
	"io"
	"net/http"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

var ccloudHandlers = []route{
	{"/api/env_metadata", handleEnvMetadata},
	{"/api/external_identities", handleExternalIdentities},
	{"/api/growth/v1/free-trial-info", handleFreeTrialInfo},
	{"/api/login/realm", handleLoginRealm},
	{"/api/metadata/security/v2alpha1/authenticate", handleV2Authenticate},
	{"/api/organizations/{id}/payment_info", handlePaymentInfo},
	{"/api/organizations/{id}/price_table", handlePriceTable},
	{"/api/organizations/{id}/promo_code_claims", handlePromoCodeClaims},
	{"/api/schema_registries", handleSchemaRegistries},
	{"/api/service_accounts", handleServiceAccounts},
	{"/api/service_accounts/{id}", handleServiceAccount},
	{"/api/sessions", handleLogin},
	{"/api/users", handleUsers},
	{"/ldapi/sdk/eval/{env}/users/{user:[a-zA-Z0-9=\\-\\/]+}", handleLaunchDarkly},
}

type CloudRouter struct {
	*mux.Router
}

// New CloudRouter with all cloud handlers
func NewCloudRouter(t *testing.T, isAuditLogEnabled bool) *CloudRouter {
	router := &CloudRouter{Router: mux.NewRouter()}
	router.Use(defaultHeaderMiddleware)

	for _, route := range ccloudHandlers {
		router.HandleFunc(route.path, route.handler(t))
	}

	router.HandleFunc("/api/me", handleMe(t, isAuditLogEnabled))
	router.addRoutesAndReplies(t, "/api/metadata/security/v2alpha1/roles", v2RoutesAndReplies)

	return router
}

func (c CloudRouter) addRoutesAndReplies(t *testing.T, base string, routesAndReplies map[string]string) {
	jsonRolesMap := rolesListToJsonMap(rbacPublicRoles())
	addRoles(base, routesAndReplies, jsonRolesMap)

	for route, reply := range routesAndReplies {
		s := reply
		c.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			_, err := io.WriteString(w, s)
			require.NoError(t, err)
		})
	}

	c.HandleFunc(base, c.HandleAllRolesRoute(t))
}
