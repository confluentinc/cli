package test_server

import (
	"fmt"
	"github.com/gorilla/mux" //https://github.com/gorilla/mux
	"net/http/httptest"
	"testing"
)

const (
	sessions			= "/api/sessions"
	me					= "/api/me"
	checkEmail			= "/api/check_email/{email}"
	account				= "/api/accounts/{id}"
	accounts			= "/api/accounts"
	acls				= "/2.0/kafka/{id}/acls"
	apiKey				= "/api/api_keys/{key}"
	apiKeys				= "/api/api_keys"
	cluster				= "/api/clusters/{id}"
	clusters			= "/api/clusters"
	envMetadata			= "/api/env_metadata"
	serviceAccounts		= "/api/service_accounts"
	schemaRegistries	= "/api/schema_registries"
	ksql				= "/api/ksqls/{id}"
	ksqls				= "/api/ksqls"
	priceTable			= "/api/organizations/{id}/price_table"
	paymentInfo			= "/api/organizations/{id}/payment_info"
	invites				= "/api/organizations/{id}/invites"
	user				= "/api/users/{id}"
	users				= "/api/users"
	connector			= "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}"
	connectorPause		= "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/pause"
	connectorResume		= "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/resume"
	connectorUpdate		= "/api/accounts/{env}/clusters/{cluster}/connectors/{connector}/config"
	connectors			= "/api/accounts/{env}/clusters/{cluster}/connectors"
	connectorPlugins	= "/api/accounts/{env}/clusters/{cluster}/connector-plugins"
	connectCatalog		= "/api/accounts/{env}/clusters/{cluster}/connector-plugins/{plugin}/config/validate"
)

var (
	apiUrl string
)

func StartTestServer(t *testing.T) *httptest.Server {
	fmt.Println("HELLLOLOLOLOLO")
	apiRouter := NewCCloudRouter(t )
	serv := httptest.NewServer(apiRouter)
	apiUrl = serv.URL
	return serv
}

func NewCCloudRouter(t *testing.T) *mux.Router {
	apiRouter := mux.NewRouter()
	buildHandler(apiRouter, t)
	return apiRouter
}

func buildHandler(router *mux.Router, t *testing.T) {
	addCcloudRoutes(router, t)
	addKafkaRoutes(router, t)
}

func addKafkaRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(acls, handleKafkaACLsList(t))
}

func addCcloudRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(sessions, handleLogin(t))
	router.HandleFunc(me, handleMe(t))
	router.HandleFunc(checkEmail, handleCheckEmail(t))
	router.HandleFunc(envMetadata, handleEnvMetadata(t))
	router.HandleFunc(serviceAccounts, handleServiceAccountRequests(t))
	router.HandleFunc(schemaRegistries, handleSchemaRegistriesRequests(t))
	addEnvironmentRoutes(router, t)
	addOrgRoutes(router, t)
	addApiKeyRoutes(router, t)
	addClusterRoutes(router, t)
	addKsqlRoutes(router, t)
	addUserRoutes(router, t)
	addConnectorsRoutes(router, t)
}

func addUserRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(user, handleUser(t))
	router.HandleFunc(users, handleUsers(t))
}

func addOrgRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(priceTable, handlePriceTable(t))
	router.HandleFunc(paymentInfo, handlePaymentInfo(t))
	router.HandleFunc(invites, handleInvite(t))
}

func addKsqlRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(ksqls, handelKsqlsRequests(t))
	router.HandleFunc(ksql, handleKsqlRequests(t))
}

func addClusterRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(clusters, handleClusters(t))
	router.HandleFunc(cluster, handleCluster(t))
}

func addApiKeyRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(apiKeys, handleApiKeys(t))
	router.HandleFunc(apiKey, handleApiKey(t))
}

func addEnvironmentRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(accounts, handleEnvironmentsRequests(t))
	router.HandleFunc(account, handleEnvironmentRequests(t))
}

func addConnectorsRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(connector, handleConnector(t))
	router.HandleFunc(connectors, handleConnectors(t))
	router.HandleFunc(connectorPause, handleConnectorPause(t))
	router.HandleFunc(connectorResume, handleConnectorResume(t))
	router.HandleFunc(connectorPlugins, handlePlugins(t))
	router.HandleFunc(connectCatalog, handleConnectCatalog(t))
	router.HandleFunc(connectorUpdate, handleConnectUpdate(t))
}
