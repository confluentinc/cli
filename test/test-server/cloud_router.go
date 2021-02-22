package test_server

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
	checkEmail          = "/api/check_email/{email}"
	account             = "/api/accounts/{id}"
	accounts            = "/api/accounts"
	apiKey              = "/api/api_keys/{key}"
	apiKeys             = "/api/api_keys"
	cluster             = "/api/clusters/{id}"
	clusters            = "/api/clusters"
	envMetadata         = "/api/env_metadata"
	serviceAccounts     = "/api/service_accounts"
	schemaRegistries    = "/api/schema_registries"
	schemaRegistry      = "/api/schema_registries/{id}"
	ksql                = "/api/ksqls/{id}"
	ksqls               = "/api/ksqls"
	priceTable          = "/api/organizations/{id}/price_table"
	paymentInfo         = "/api/organizations/{id}/payment_info"
	invites             = "/api/organizations/{id}/invites"
	user                = "/api/users/{id}"
	users               = "/api/users"
	connector           = "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}"
	connectorPause      = "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/pause"
	connectorResume     = "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/resume"
	connectorUpdate     = "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/config"
	connectors          = "/api/accounts/{env}/clusters/{cluster}/connectors"
	connectorPlugins    = "/api/accounts/{env}/clusters/{cluster}/connector-plugins"
	connectCatalog      = "/api/accounts/{env}/clusters/{cluster}/connector-plugins/{plugin}/config/validate"
	v2alphaAuthenticate = "/api/metadata/security/v2alpha1/authenticate"
	signup              = "/api/signup"
	verifyEmail         = "/api/email_verifications"
)

type CloudRouter struct {
	*mux.Router
	kafkaApiUrl string
	srApiUrl    string
	kafkaRPUrl  string
}

// New CloudRouter with all cloud handlers
func NewCloudRouter(t *testing.T) *CloudRouter {
	c := NewEmptyCloudRouter()
	c.buildCcloudRouter(t)
	return c
}

// New CLoudRouter with no predefined handlers
func NewEmptyCloudRouter() *CloudRouter {
	return &CloudRouter{
		Router: mux.NewRouter(),
	}
}

// Add handlers for cloud endpoints
func (c *CloudRouter) buildCcloudRouter(t *testing.T) {
	c.HandleFunc(sessions, c.HandleLogin(t))
	c.HandleFunc(me, c.HandleMe(t))
	c.HandleFunc(checkEmail, c.HandleCheckEmail(t))
	c.HandleFunc(signup, c.HandleSignup(t))
	c.HandleFunc(verifyEmail, c.HandleSendVerificationEmail(t))
	c.HandleFunc(envMetadata, c.HandleEnvMetadata(t))
	c.HandleFunc(serviceAccounts, c.HandleServiceAccount(t))
	c.addSchemaRegistryRoutes(t)
	c.addEnvironmentRoutes(t)
	c.addOrgRoutes(t)
	c.addApiKeyRoutes(t)
	c.addClusterRoutes(t)
	c.addKsqlRoutes(t)
	c.addUserRoutes(t)
	c.addConnectorsRoutes(t)
	c.addV2AlphaRoutes(t)
}

func (c CloudRouter) addV2AlphaRoutes(t *testing.T) {
	c.HandleFunc(v2alphaAuthenticate, c.HandleV2Authenticate(t))
	c.addRoutesAndReplies(t, v2Base, v2RoutesAndReplies, v2RbacRoles)
}

func (c CloudRouter) addRoutesAndReplies(t *testing.T, base string, routesAndReplies, rbacRoles map[string]string) {
	addRoles(base, routesAndReplies, rbacRoles)
	for route, reply := range routesAndReplies {
		s := reply
		c.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			require.NoError(t, err)
		})
	}
}

func (c *CloudRouter) addSchemaRegistryRoutes(t *testing.T) {
	c.HandleFunc(schemaRegistries, c.HandleSchemaRegistries(t))
	c.HandleFunc(schemaRegistry, c.HandleSchemaRegistry(t))
}

func (c *CloudRouter) addUserRoutes(t *testing.T) {
	c.HandleFunc(user, c.HandleUser(t))
	c.HandleFunc(users, c.HandleUsers(t))
}

func (c *CloudRouter) addOrgRoutes(t *testing.T) {
	c.HandleFunc(priceTable, c.HandlePriceTable(t))
	c.HandleFunc(paymentInfo, c.HandlePaymentInfo(t))
	c.HandleFunc(invites, c.HandleInvite(t))
}

func (c *CloudRouter) addKsqlRoutes(t *testing.T) {
	c.HandleFunc(ksqls, c.HandleKsqls(t))
	c.HandleFunc(ksql, c.HandleKsql(t))
}

func (c *CloudRouter) addClusterRoutes(t *testing.T) {
	c.HandleFunc(clusters, c.HandleClusters(t))
	c.HandleFunc(cluster, c.HandleCluster(t))
}

func (c *CloudRouter) addApiKeyRoutes(t *testing.T) {
	c.HandleFunc(apiKeys, c.HandleApiKeys(t))
	c.HandleFunc(apiKey, c.HandleApiKey(t))
}

func (c *CloudRouter) addEnvironmentRoutes(t *testing.T) {
	c.HandleFunc(accounts, c.HandleEnvironments(t))
	c.HandleFunc(account, c.HandleEnvironment(t))
}

func (c *CloudRouter) addConnectorsRoutes(t *testing.T) {
	c.HandleFunc(connector, c.HandleConnector(t))
	c.HandleFunc(connectors, c.HandleConnectors(t))
	c.HandleFunc(connectorPause, c.HandleConnectorPause(t))
	c.HandleFunc(connectorResume, c.HandleConnectorResume(t))
	c.HandleFunc(connectorPlugins, c.HandlePlugins(t))
	c.HandleFunc(connectCatalog, c.HandleConnectCatalog(t))
	c.HandleFunc(connectorUpdate, c.HandleConnectUpdate(t))
}
