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
	"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:sa-12345": `[
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
	"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-55eee": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595",
						"cloud-cluster=lkc-1111aaa"
					],
					"clusters": {
						"kafka-cluster": "lkc-1111aaa"
					}
				},
				"rolebindings": {
					"User:u-55eee": {
						"ResourceOwner": [
							{ "resourceType": "Topic", "name": "clicks-", "patternType": "PREFIX" },
							{ "resourceType": "Topic", "name": "payroll", "patternType": "LITERAL" },
							{ "resourceType": "Group", "name": "readers", "patternType": "LITERAL" }
						]
					}
				}
		  	}
		]`,
	"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-66fff": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595",
						"cloud-cluster=lkc-1111aaa"
					],
					"clusters": {
						"ksql-cluster": "ksql-cluster-name-2222bbb"
					}
				},
				"rolebindings": {
					"User:u-66fff": {
						"ResourceOwner": [
							{ "resourceType": "Cluster", "name": "ksql-cluster-name-2222bbb", "patternType": "LITERAL" }
						]
					}
				}
		  	}
		]`,
	"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-66ffa": `[
		  	{
				"scope": {
				  	"path": [
						"organization=2345",
						"environment=b-595",
						"cloud-cluster=lkc-1234abc"
					],
					"clusters": {
						"ksql-cluster": "ksqlDB_cluster_name"
					}
				},
				"rolebindings": {
					"User:u-66ffa": {
						"KsqlAdmin": []
					}
				}
		  	}
		]`,
	"/api/metadata/security/v2alpha1/lookup/rolebindings/principal/User:u-77ggg": `[
		  	{
				"scope": {
				  	"path": [
						"organization=1234",
						"environment=a-595"
					],
					"clusters": {
						"schema-registry-cluster": "lsrc-3333ccc"
					}
				},
				"rolebindings": {
					"User:u-66fff": {
						"ResourceOwner": [
							{ "resourceType": "Subject", "name": "clicks", "patternType": "LITERAL" }
						]
					}
				}
		  	}
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
