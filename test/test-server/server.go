package test_server

import (
	"fmt"
	"github.com/gorilla/mux" //https://github.com/gorilla/mux
	"github.com/stretchr/testify/require"
	"io"
	"encoding/json"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"net/http"
	"net/http/httptest"
	"path"
	"reflect"
	"sort"
	"strings"
	"testing"
)

const (
	sessions			= "/api/sessions"
	me					= "/api/me"
	checkEmail			= "/api/check_email/{email}"
	account				= "/api/accounts/{id}"
	accounts			= "/api/accounts"
	apiKey				= "/api/api_keys/{key}"
	apiKeys				= "/api/api_keys"
	cluster				= "/api/clusters/{id}"
	clusters			= "/api/clusters"
	envMetadata			= "/api/env_metadata"
	serviceAccounts		= "/api/service_accounts"
	schemaRegistries	= "/api/schema_registries"
	schemaRegistry		= "/api/schema_registries/{id}"
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

type CloudRouter struct {
	*mux.Router
	kafkaApiUrl string
}

type KafkaRouter struct {
	*mux.Router
}
type CloudTestBackend struct {
	cloud *httptest.Server
	cloudRouter *CloudRouter
	kafka *httptest.Server
	kafkaRouter *KafkaRouter
}

func StartTestBackend(t *testing.T) *CloudTestBackend {
	//apiRouter := NewEmptyRouter()
	//cloudServ := httptest.NewServer(apiRouter)
	//apiRouter := NewCCloudRouter(t)
	//cloudServ := httptest.NewServer(apiRouter)
	cloudRouter := NewCCloudRouter(t)
	kafkaRouter := NewKafkaRouter(t)
	//kafkaServ := httptest.NewServer(kafkaRouter)
	//cloudRouter := &CloudRouter{
	//	Router:      apiRouter,
	//	//kafkaApiUrl: kafkaServ.URL,
	//}
	ccloud := &CloudTestBackend{
		cloud:	httptest.NewServer(cloudRouter),
		cloudRouter: cloudRouter,
		kafka:	httptest.NewServer(kafkaRouter),
		kafkaRouter: kafkaRouter,
	}
	cloudRouter.kafkaApiUrl = ccloud.kafka.URL
	//cloudRouter.buildCloudHandler(cloudRouter.Router, t)
	//kafkaRouter.buildKafkaHandler(kafkaRouter.Router, t)
	return ccloud
}

func (b *CloudTestBackend) Close() {
	b.cloud.Close()
	b.kafka.Close()
}

func (b *CloudTestBackend) GetCloudUrl() string{
	return b.cloud.URL
}

func (b *CloudTestBackend) GetKafkaUrl() string{
	return b.kafka.URL
}

func NewSingleTestBackend(cloudRouter *CloudRouter, kafkaRouter *KafkaRouter) *CloudTestBackend {
	ccloud := &CloudTestBackend{
		cloud:	httptest.NewServer(cloudRouter),
		cloudRouter: cloudRouter,
		kafka:	httptest.NewServer(kafkaRouter),
		kafkaRouter: kafkaRouter,
	}
	ccloud.cloudRouter.kafkaApiUrl = ccloud.kafka.URL
	return ccloud
}

func NewCCloudRouter(t *testing.T) *CloudRouter {
	c := NewEmptyCloudRouter()//NewEmptyRouter()
	//c := &CloudRouter{
	//	Router:      apiRouter,
	//}
	c.buildCloudHandler(c.Router, t)
	return c
}

func NewEmptyCloudRouter() *CloudRouter {
	return &CloudRouter{
		Router: mux.NewRouter(),
	}
}

func NewKafkaRouter(t *testing.T) *KafkaRouter {
	router := NewEmptyKafkaRouter()
	router.buildKafkaHandler(router.Router, t)
	return router
}

func NewEmptyKafkaRouter() *KafkaRouter {
	return &KafkaRouter{
		mux.NewRouter(),
	}
}

func (c *CloudRouter) GetCloudHandler(t *testing.T, name string) func(w http.ResponseWriter, r *http.Request) {
	fmt.Println("trying to get handler")
	fmt.Println(name)
	m := reflect.ValueOf(c).MethodByName(name)
	fmt.Println(m)
	in := []reflect.Value{reflect.ValueOf(t)}
	handlerFunc := m.Call(in)
	return handlerFunc[0].Interface().(func(w http.ResponseWriter, r *http.Request))
}

func (c *CloudRouter) buildCloudHandler(router *mux.Router, t *testing.T) {
	c.addCcloudRoutes(router, t)
}

// kafka urls
const (
	aclsCreate				= "/2.0/kafka/{id}/acls"
	aclsList				= "/2.0/kafka/{cluster}/acls:search"
	aclsDelete				= "/2.0/kafka/{cluster}/acls/delete"
	link					= "/2.0/kafka/{cluster}/links/{link}"
	links					= "/2.0/kafka/{cluster}/links"
	topicMirrorStop			= "/2.0/kafka/{cluster}/topics/{topic}/mirror:stop"
)
func (c *KafkaRouter) buildKafkaHandler(router *mux.Router, t *testing.T) {
	router.HandleFunc(aclsCreate, c.HandleKafkaACLsCreate(t))
	router.HandleFunc(aclsList, c.HandleKafkaACLsList(t))
	router.HandleFunc(aclsDelete, c.HandleKafkaACLsDelete(t))
	router.HandleFunc(link, c.HandleKafkaLink(t))
	router.HandleFunc(links, c.HandleKafkaLinks(t))
	router.HandleFunc(topicMirrorStop, c.HandleKafkaTopicMirrorStop(t))
	router.PathPrefix("/").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, err := io.WriteString(w, `{}`)
		require.NoError(t, err)
	})
}

func (c *CloudRouter) addCcloudRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(sessions, c.HandleLogin(t))
	router.HandleFunc(me, c.HandleMe(t))
	router.HandleFunc(checkEmail, c.HandleCheckEmail(t))
	router.HandleFunc(envMetadata, c.HandleEnvMetadata(t))
	router.HandleFunc(serviceAccounts, c.HandleServiceAccountRequests(t))
	c.addSchemaRegistryRoutes(router, t)
	c.addEnvironmentRoutes(router, t)
	c.addOrgRoutes(router, t)
	c.addApiKeyRoutes(router, t)
	c.addClusterRoutes(router, t)
	c.addKsqlRoutes(router, t)
	c.addUserRoutes(router, t)
	c.addConnectorsRoutes(router, t)
	addMdsv2alpha1(t, router)
}

func (c *CloudRouter) addSchemaRegistryRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(schemaRegistries, c.HandleSchemaRegistriesRequests(t))
	router.HandleFunc(schemaRegistry, c.HandleSchemaRegistriesRequests(t))
}

func (c *CloudRouter) addUserRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(user, c.HandleUser(t))
	router.HandleFunc(users, c.HandleUsers(t))
}

func (c *CloudRouter) addOrgRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(priceTable, c.HandlePriceTable(t))
	router.HandleFunc(paymentInfo, c.HandlePaymentInfo(t))
	router.HandleFunc(invites, c.HandleInvite(t))
}

func (c *CloudRouter) addKsqlRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(ksqls, c.HandleKsqlsRequests(t))
	router.HandleFunc(ksql, c.HandleKsqlRequests(t))
}

func (c *CloudRouter) addClusterRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(clusters, c.HandleClusters(t))
	router.HandleFunc(cluster, c.HandleCluster(t))
}

func (c *CloudRouter) addApiKeyRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(apiKeys, c.HandleApiKeys(t))
	router.HandleFunc(apiKey, c.HandleApiKey(t))
}

func (c *CloudRouter) addEnvironmentRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(accounts, c.HandleEnvironmentsRequests(t))
	router.HandleFunc(account, c.HandleEnvironmentRequests(t))
}

func (c *CloudRouter) addConnectorsRoutes(router *mux.Router, t *testing.T) {
	router.HandleFunc(connector, c.HandleConnector(t))
	router.HandleFunc(connectors, c.HandleConnectors(t))
	router.HandleFunc(connectorPause, c.HandleConnectorPause(t))
	router.HandleFunc(connectorResume, c.HandleConnectorResume(t))
	router.HandleFunc(connectorPlugins, c.HandlePlugins(t))
	router.HandleFunc(connectCatalog, c.HandleConnectCatalog(t))
	router.HandleFunc(connectorUpdate, c.HandleConnectUpdate(t))
}

func addMdsv2alpha1(t *testing.T, router *mux.Router) {
	req := require.New(t)
	router.HandleFunc("/api/metadata/security/v2alpha1/authenticate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		reply := &mds.AuthenticationResponse{
			AuthToken: "eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpc3MiOiJPbmxpbmUgSldUIEJ1aWxkZXIiLCJpYXQiOjE1NjE2NjA4NTcsImV4cCI6MjUzMzg2MDM4NDU3LCJhdWQiOiJ3d3cuZXhhbXBsZS5jb20iLCJzdWIiOiJqcm9ja2V0QGV4YW1wbGUuY29tIn0.G6IgrFm5i0mN7Lz9tkZQ2tZvuZ2U7HKnvxMuZAooPmE",
			TokenType: "dunno",
			ExpiresIn: 9999999999,
		}
		b, err := json.Marshal(&reply)
		req.NoError(err)
		_, err = io.WriteString(w, string(b))
		req.NoError(err)
	})
	router.Handle("/api/metadata/security/v2alpha1/roles/InvalidRole", http.NotFoundHandler())

	routesAndReplies := map[string]string{
		"/api/metadata/security/v2alpha1/principals/User:u-11aaa/roles/CloudClusterAdmin": `[]`,
		"/api/metadata/security/v2alpha1/roleNames": `[
			"CCloudRoleBindingAdmin",
			"CloudClusterAdmin",
			"EnvironmentAdmin",
			"OrganizationAdmin"
		]`,
		"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-11aaa": `[
			{
				"scope": {
				  	"path": [
						"organization=1111aaaa-11aa-11aa-11aa-111111aaaaaac"
					],
					"clusters": {
					}
				},
				"rolebindings": {
					"User:u-11aaa": {
						"OrganizationAdmin": []
					}
				}
		  	},
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595",
						"cloud-cluster=lkc-1111aaa"
					],
					"clusters": {
					}
				},
				"rolebindings": {
					"User:u-11aaa": {
						"CloudClusterAdmin": []
					}
				}
		  	}
		]`,
		"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-22bbb": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595"
					],
					"clusters": {
					}
				},
				"rolebindings": {
					"User:u-22bbb": {
						"EnvironmentAdmin": []
					}
				}
		  	}
		]`,
		"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-33ccc": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595",
						"cloud-cluster=lkc-1111aaa"
					],
					"clusters": {
					}
				},
				"rolebindings": {
					"User:u-33ccc": {
						"CloudClusterAdmin": []
					}
				}
		  	}
		]`,
		"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-44ddd": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595",
						"cloud-cluster=lkc-1111aaa"
					],
					"clusters": {
					}
				},
				"rolebindings": {
					"User:u-44ddd": {
						"CloudClusterAdmin": []
					}
				}
		  	}
		]`,
		"/api/metadata/security/v2alpha1/lookup/role/OrganizationAdmin": `[
			"User:u-11aaa"
		]`,
		"/api/metadata/security/v2alpha1/lookup/role/EnvironmentAdmin": `[
			"User:u-22bbb"
		]`,
		"/api/metadata/security/v2alpha1/lookup/role/CloudClusterAdmin": `[
			"User:u-33ccc", "User:u-44ddd"
		]`,
	}
	addRolesV2(routesAndReplies)

	for route, reply := range routesAndReplies {
		s := reply
		router.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			req.NoError(err)
		})
	}
}

var (
	rbacRolesV2 = map[string]string{
		"CCloudRoleBindingAdmin": `{
			"name": "CCloudRoleBindingAdmin",
			"policy": {
				"bindingScope": "root",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType":"SecurityMetadata","operations":["Describe","Alter"]},
				{"resourceType":"Organization","operations":["AlterAccess","DescribeAccess"]}]}}`,
		"CloudClusterAdmin": `{
			"name": "CloudClusterAdmin",
			"policies": [
			{
				"bindingScope": "cluster",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType": "Topic","operations": ["All"]},
				{"resourceType": "KsqlCluster","operations": ["All"]},
				{"resourceType": "Subject","operations": ["All"]},
				{"resourceType": "Connector","operations": ["All"]},
				{"resourceType": "NetworkAccess","operations": ["All"]},
				{"resourceType": "ClusterMetric","operations": ["All"]},
				{"resourceType": "Cluster","operations": ["All"]},
				{"resourceType": "ClusterApiKey","operations": ["All"]},
				{"resourceType": "SecurityMetadata","operations": ["Describe", "Alter"]}]
			},
			{
				"bindingScope": "organization",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType": "SupportPlan","operations": ["Describe"]},
				{"resourceType": "User","operations": ["Describe","Invite"]},
				{"resourceType": "ServiceAccount","operations": ["Describe"]}]
			}]}`,
		"EnvironmentAdmin": `{
			"name": "EnvironmentAdmin",
			"policies": [
			{
				"bindingScope": "ENVIRONMENT",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType": "SecurityMetadata","operations": ["Describe", "Alter"]},
				{"resourceType": "ClusterApiKey","operations": ["All"]},
				{"resourceType": "Connector","operations": ["All"]},
				{"resourceType": "NetworkAccess","operations": ["All"]},
				{"resourceType": "KsqlCluster","operations": ["All"]},
				{"resourceType": "Environment","operations": ["Alter","Delete","AlterAccess","CreateKafkaCluster","DescribeAccess"]},
				{"resourceType": "Subject","operations": ["All"]},
				{"resourceType": "NetworkConfig","operations": ["All"]},
				{"resourceType": "ClusterMetric","operations": ["All"]},
				{"resourceType": "Cluster","operations": ["All"]},
				{"resourceType": "SchemaRegistry","operations": ["All"]},
				{"resourceType": "NetworkRegion","operations": ["All"]},
				{"resourceType": "Deployment","operations": ["All"]},
				{"resourceType": "Topic","operations": ["All"]}
				]
			},
			{
				"bindingScope": "organization",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType": "User","operations": ["Describe","Invite"]},
				{"resourceType": "ServiceAccount","operations": ["Describe"]},
				{"resourceType": "SupportPlan","operations": ["Describe"]}
				]
			}]}`,
		"OrganizationAdmin": `{
			"name": "OrganizationAdmin",
			"policy": {
				"bindingScope": "organization",
				"bindWithResource": false,
				"allowedOperations": [
				{"resourceType": "Topic","operations": ["All"]},
				{"resourceType": "NetworkConfig","operations": ["All"]},
				{"resourceType": "SecurityMetadata","operations": ["Describe", "Alter"]},
				{"resourceType": "Billing","operations": ["All"]},
				{"resourceType": "ClusterApiKey","operations": ["All"]},
				{"resourceType": "Deployment","operations": ["All"]},
				{"resourceType": "SchemaRegistry","operations": ["All"]},
				{"resourceType": "KsqlCluster","operations": ["All"]},
				{"resourceType": "CloudApiKey","operations": ["All"]},
				{"resourceType": "NetworkAccess","operations": ["All"]},
				{"resourceType": "SecuritySSO","operations": ["All"]},
				{"resourceType": "SupportPlan","operations": ["All"]},
				{"resourceType": "Connector","operations": ["All"]},
				{"resourceType": "ClusterMetric","operations": ["All"]},
				{"resourceType": "ServiceAccount","operations": ["All"]},
				{"resourceType": "Subject","operations": ["All"]},
				{"resourceType": "Cluster","operations": ["All"]},
				{"resourceType": "Environment","operations": ["All"]},
				{"resourceType": "NetworkRegion","operations": ["All"]},
				{"resourceType": "Organization","operations": ["Alter","CreateEnvironment","AlterAccess","DescribeAccess"]},
				{"resourceType": "User","operations": ["All"]}
				]
			}
		}`,
	}
)

func addRolesV2(routesAndReplies map[string]string) {
	base := "/api/metadata/security/v2alpha1/roles"
	var roleNameList []string
	for roleName, roleInfo := range rbacRolesV2 {
		routesAndReplies[path.Join(base, roleName)] = roleInfo
		roleNameList = append(roleNameList, roleName)
	}

	sort.Strings(roleNameList)

	var allRoles []string
	for _, roleName := range roleNameList {
		allRoles = append(allRoles, rbacRolesV2[roleName])
	}
	routesAndReplies[base] = "[" + strings.Join(allRoles, ",") + "]"
}
