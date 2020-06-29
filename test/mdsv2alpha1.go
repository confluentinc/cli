package test

import (
	"encoding/json"
	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
	"io"
	"net/http"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

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

func addMdsv2alpha1(t *testing.T, router *http.ServeMux) {
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

	routesAndReplies := map[string]string{
		"/api/metadata/security/v2alpha1/lookup/managed/rolebindings/principal/User:u-11111": `[
			{
				"scope": {
				  	"path": [
						"organization=1234"
					],
					"clusters": {
					}
				},
				"cluster_role_bindings": {
					"User:u-11111": {
						"OrganizationAdmin": [
							{
								"resourceType": "Organization",
								"name": "org1",
								"patternType": "LITERAL"
							}
						]
					}
				},
				"resource_role_bindings": {}
		  	},
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595"
					],
					"clusters": {
						"kafka-cluster": "cluster1"
					}
				},
				"cluster_role_bindings": {
					"User:u-11111": {
						"CloudClusterAdmin": [
							{
								"resourceType": "Cluster",
								"name": "cluster1",
								"patternType": "LITERAL"
							}
						]
					},
					"User:u-22222": {
						"EnvironmentAdmin": [
							{
								"resourceType": "Environment",
								"name": "env1",
								"patternType": "LITERAL"
						  	}
						]
					},
					"User:u-33333": {
						"CloudClusterAdmin": [
							{
								"resourceType": "Cluster",
								"name": "cluster1",
								"patternType": "LITERAL"
						  	}
						]
					},
					"User:u-44444": {
						"CloudClusterAdmin": [
							{
								"resourceType": "Cluster",
								"name": "cluster1",
								"patternType": "LITERAL"
						  	}
						]
					}
				},
				"resource_role_bindings": {}
		  	}
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

func addRolesV2(routesAndReplies map[string]string) {
	base := "/api/metadata/security/v2alpha1/roles"
	var roleNameList []string
	for roleName, roleInfo := range rbacRolesV2 {
		routesAndReplies[base+"/"+roleName] = roleInfo
		roleNameList = append(roleNameList, roleName)
	}

	sort.Strings(roleNameList)

	var allRoles []string
	for _, roleName := range roleNameList {
		allRoles = append(allRoles, rbacRolesV2[roleName])
	}
	routesAndReplies[base] = "[" + strings.Join(allRoles, ",") + "]"
}

/*
func handleRegistryClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/json")
		clusterType := r.URL.Query().Get("clusterType")
		response := `[ {
		"name": "theMdsConnectCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "connect-cluster": "connect-name" } },
		"hosts": [ { "host": "10.5.5.5", "port": 9005 } ]
	  },{
		"name": "theMdsKafkaCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID" } },
		"hosts": [ { "host": "10.10.10.10", "port": 8090 },{ "host": "mds.example.com", "port": 8090 } ]
	  },{
		"name": "theMdsKSQLCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		"hosts": [ { "host": "10.4.4.4", "port": 9004 } ]
	  },{
		"name": "theMdsSchemaRegistryCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "schema-registry-cluster": "schema-registry-name" } },
		"hosts": [ { "host": "10.3.3.3", "port": 9003 } ]
	} ]`
		if clusterType == "ksql-cluster" {
			response = `[ {
		    "name": "theMdsKSQLCluster",
		    "scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		    "hosts": [ { "host": "10.4.4.4", "port": 9004 } ]
		  } ]`
		}
		_, err := io.WriteString(w, response)
		require.NoError(t, err)
	}
}
 */
