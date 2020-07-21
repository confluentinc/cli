package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	mds "github.com/confluentinc/mds-sdk-go/mdsv1"
)

var (
	rbacRoles = map[string]string{
		"DeveloperRead": `{
                      "name":"DeveloperRead",
                      "accessPolicy":{
                              "scopeType":"Resource",
                              "allowedOperations":[
                                      {"resourceType":"Cluster","operations":[]},
                                      {"resourceType":"TransactionalId","operations":["Describe"]},
                                      {"resourceType":"Group","operations":["Read","Describe"]},
                                      {"resourceType":"Subject","operations":["Read","ReadCompatibility"]},
                                      {"resourceType":"Connector","operations":["ReadStatus","ReadConfig"]},
                                      {"resourceType":"Topic","operations":["Read","Describe"]}]}}`,
		"DeveloperWrite": `{
                      "name":"DeveloperWrite",
                      "accessPolicy":{
                              "scopeType":"Resource",
                              "allowedOperations":[
                                      {"resourceType":"Subject","operations":["Write"]},
                                      {"resourceType":"Group","operations":[]},
                                      {"resourceType":"Topic","operations":["Write","Describe"]},
                                      {"resourceType":"Cluster","operations":["IdempotentWrite"]},
                                      {"resourceType":"KsqlCluster","operations":["Contribute"]},
                                      {"resourceType":"Connector","operations":["ReadStatus","Configure"]},
                                      {"resourceType":"TransactionalId","operations":["Write","Describe"]}]}}`,
		"SecurityAdmin": `{
                      "name":"SecurityAdmin",
                      "accessPolicy":{
                              "scopeType":"Cluster",
                              "allowedOperations":[
                                      {"resourceType":"All","operations":["DescribeAccess"]}]}}`,
		"SystemAdmin": `{
                      "name":"SystemAdmin",
                      "accessPolicy":{
                              "scopeType":"Cluster",
                              "allowedOperations":[
                                      {"resourceType":"All","operations":["All"]}]}}`,
	}
)

func serveMds(t *testing.T) *httptest.Server {
	req := require.New(t)
	router := http.NewServeMux()
	router.HandleFunc("/security/1.0/authenticate", func(w http.ResponseWriter, r *http.Request) {
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
		"/security/1.0/principals/User:frodo/groups": `[
                       "hobbits",
                       "ringBearers"]`,
		"/security/1.0/principals/User:frodo/roleNames": `[
                       "DeveloperRead",
                       "DeveloperWrite",
                       "SecurityAdmin"]`,
		"/security/1.0/principals/User:frodo/roles/DeveloperRead/resources":  `[]`,
		"/security/1.0/principals/User:frodo/roles/DeveloperWrite/resources": `[]`,
		"/security/1.0/principals/User:frodo/roles/SecurityAdmin/resources":  `[]`,
		"/security/1.0/principals/Group:hobbits/roles/DeveloperRead/resources": `[
                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]`,
		"/security/1.0/principals/Group:hobbits/roles/DeveloperWrite/resources": `[
                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}]`,
		"/security/1.0/principals/Group:hobbits/roles/SecurityAdmin/resources":     `[]`,
		"/security/1.0/principals/Group:ringBearers/roles/DeveloperRead/resources": `[]`,
		"/security/1.0/principals/Group:ringBearers/roles/DeveloperWrite/resources": `[
                       {"resourceType":"Topic","name":"ring-","patternType":"PREFIXED"}]`,
		"/security/1.0/principals/Group:ringBearers/roles/SecurityAdmin/resources": `[]`,
		"/security/1.0/lookup/principal/User:frodo/resources": `{
                       "Group:hobbits":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}],
                               "DeveloperRead":[
                                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]},
                       "Group:ringBearers":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"ring-","patternType":"PREFIXED"}]},
                       "User:frodo":{
                               "SecurityAdmin": []}}`,
		"/security/1.0/lookup/principal/Group:hobbits/resources": `{
                       "Group:hobbits":{
                               "DeveloperWrite":[
                                       {"resourceType":"Topic","name":"shire-","patternType":"PREFIXED"}],
                               "DeveloperRead":[
                                       {"resourceType":"Topic","name":"drink","patternType":"LITERAL"},
                                       {"resourceType":"Topic","name":"food","patternType":"LITERAL"}]}}`,
		"/security/1.0/lookup/role/DeveloperRead":                                    `["Group:hobbits"]`,
		"/security/1.0/lookup/role/DeveloperWrite":                                   `["Group:hobbits","Group:ringBearers"]`,
		"/security/1.0/lookup/role/SecurityAdmin":                                    `["User:frodo"]`,
		"/security/1.0/lookup/role/SystemAdmin":                                      `[]`,
		"/security/1.0/lookup/role/DeveloperRead/resource/Topic/name/food":           `["Group:hobbits"]`,
		"/security/1.0/lookup/role/DeveloperRead/resource/Topic/name/shire-parties":  `[]`,
		"/security/1.0/lookup/role/DeveloperWrite/resource/Topic/name/shire-parties": `["Group:hobbits"]`,
	}
	addRoles(routesAndReplies)

	for route, reply := range routesAndReplies {
		s := reply
		router.HandleFunc(route, func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/json")
			_, err := io.WriteString(w, s)
			req.NoError(err)
		})
	}

	router.HandleFunc("/security/1.0/registry/clusters", handleRegistryClusters(t))

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to mds `+r.URL.Path+`"}}`)
		require.NoError(t, err)
	})
	return httptest.NewServer(router)
}

func addRoles(routesAndReplies map[string]string) {
	base := "/security/1.0/roles"
	var roleNameList []string
	for roleName, roleInfo := range rbacRoles {
		routesAndReplies[path.Join(base, roleName)] = roleInfo
		roleNameList = append(roleNameList, roleName)
	}

	sort.Strings(roleNameList)

	var allRoles []string
	for _, roleName := range roleNameList {
		allRoles = append(allRoles, rbacRoles[roleName])
	}
	routesAndReplies[base] = "[" + strings.Join(allRoles, ",") + "]"
}

func handleRegistryClusters(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			w.Header().Set("Content-Type", "text/json")
			clusterType := r.URL.Query().Get("clusterType")
			response := `[ {
		"clusterName": "theMdsConnectCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "connect-cluster": "connect-name" } },
		"hosts": [ { "host": "10.5.5.5", "port": 9005 } ],
        "protocol": "HTTPS"
	  },{
		"clusterName": "theMdsKSQLCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		"hosts": [ { "host": "10.4.4.4", "port": 9004 } ],
        "protocol": "HTTPS"
	  },{
		"clusterName": "theMdsKafkaCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID" } },
		"hosts": [ { "host": "10.10.10.10", "port": 8090 },{ "host": "mds.example.com", "port": 8090 } ],
        "protocol": "SASL_PLAINTEXT"
	  },{
		"clusterName": "theMdsSchemaRegistryCluster",
		"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "schema-registry-cluster": "schema-registry-name" } },
		"hosts": [ { "host": "10.3.3.3", "port": 9003 } ],
        "protocol": "HTTPS"
	} ]`
			if clusterType == "ksql-cluster" {
				response = `[ {
		    "clusterName": "theMdsKSQLCluster",
		    "scope": { "clusters": { "kafka-cluster": "kafka-GUID", "ksql-cluster": "ksql-name" } },
		    "hosts": [ { "host": "10.4.4.4", "port": 9004 } ],
            "protocol": "HTTPS"
			} ]`
			}
			if clusterType == "kafka-cluster" {
				response = `[ {
			"clusterName": "theMdsKafkaCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID" } },
			"hosts": [ { "host": "10.10.10.10", "port": 8090 },{ "host": "mds.example.com", "port": 8090 } ],
        	"protocol": "SASL_PLAINTEXT"
			} ]`
			}
			if clusterType == "connect-cluster" {
				response = `[ {
			"clusterName": "theMdsConnectCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "connect-cluster": "connect-name" } },
			"hosts": [ { "host": "10.5.5.5", "port": 9005 } ],
        	"protocol": "HTTPS"
			} ]`
			}
			if clusterType == "schema-registry-cluster" {
				response = `[ {
			"clusterName": "theMdsSchemaRegistryCluster",
			"scope": { "clusters": { "kafka-cluster": "kafka-GUID", "schema-registry-cluster": "schema-registry-name" } },
			"hosts": [ { "host": "10.3.3.3", "port": 9003 } ],
        	"protocol": "HTTPS"
			} ]`
			}
			_, err := io.WriteString(w, response)
			require.NoError(t, err)
		}

		if r.Method == "DELETE" {
			clusterName := r.URL.Query().Get("clusterName")
			require.NotEmpty(t, clusterName)
		}

		if r.Method == "POST" {
			var clusterInfos []*mds.ClusterInfo
			err := json.NewDecoder(r.Body).Decode(&clusterInfos)
			require.NoError(t, err)
			require.NotEmpty(t, clusterInfos)
			for _, clusterInfo := range clusterInfos {
				require.NotEmpty(t, clusterInfo.ClusterName)
				require.NotEmpty(t, clusterInfo.Hosts)
				require.NotEmpty(t, clusterInfo.Scope)
				require.NotEmpty(t, clusterInfo.Protocol)
			}
		}
	}
}
