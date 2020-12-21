package test_server

var v1RbacRoles = map[string]string{
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
var v1RoutesAndReplies = map[string]string{
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
