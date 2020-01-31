package test

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/mds-sdk-go"
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

func serveMds(t *testing.T, mdsURL string) *httptest.Server {
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

	router.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, err := io.WriteString(w, `{"error": {"message": "unexpected call to `+r.URL.Path+`"}}`)
		require.NoError(t, err)
	})
	return httptest.NewServer(router)
}

func addRoles(routesAndReplies map[string]string) {
	base := "/security/1.0/roles"
	var roleNameList []string
	for roleName, roleInfo := range rbacRoles {
		routesAndReplies[base + "/" + roleName] = roleInfo
		roleNameList = append(roleNameList, roleName)
	}

	sort.Strings(roleNameList)
	
	var allRoles []string
	for _, roleName := range roleNameList {
		allRoles = append(allRoles, rbacRoles[roleName])
	}
	routesAndReplies[base] = "[" + strings.Join(allRoles, ",") + "]"
}


