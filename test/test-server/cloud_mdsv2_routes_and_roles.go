package testserver

var v2RoutesAndReplies = map[string]string{
	"/api/metadata/security/v2alpha1/principals/User:u-11aaa/roles/CloudClusterAdmin":      `[]`,
	"/api/metadata/security/v2alpha1/principals/User:u-11aaa/roles/ResourceOwner/bindings": `[]`,
	"/api/metadata/security/v2alpha1/roleNames": `[
			"CCloudRoleBindingAdmin",
			"CloudClusterAdmin",
			"EnvironmentAdmin",
			"OrganizationAdmin"
		]`,
	"/api/metadata/security/v2alpha1/lookup/role/OrganizationAdmin": `[
			"User:u-11aaa",
			"User:sa-12345",
			"User:pool-12345"
		]`,
	"/api/metadata/security/v2alpha1/lookup/role/EnvironmentAdmin": `[
			"User:u-22bbb"
		]`,
	"/api/metadata/security/v2alpha1/lookup/role/CloudClusterAdmin": `[
			"User:u-33ccc", "User:u-44ddd", "User:u-unlisted"
		]`,
	"/api/metadata/security/v2alpha1/lookup/role/ResourceOwner/resource/Topic/name/food":          `["User:u-11aaa"]`,
	"/api/metadata/security/v2alpha1/lookup/role/ResourceOwner/resource/Topic/name/shire-parties": `["User:u-11aaa"]`,
	"/api/metadata/security/v2alpha1/lookup/role/InvalidOrgAdmin":                                 `{"status_code":400, "message":"Invalid role name : InvalidOrgAdmin","type":"INVALID REQUEST DATA"}`,
	"/api/metadata/security/v2alpha1/lookup/role/InvalidMetricsViewer": `{
	  "errors": [
		{
		  "id": "806263f440c3de28bd94fb6d1d81ac1b",
		  "status": "400",
		  "code": "invalid-role",
		  "detail": "Invalid role name : InvalidMetricsViewer",
		  "source": {}
		}
	  ]
	}`,
}
