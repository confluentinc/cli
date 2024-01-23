ALL_SRC = $(shell find . -name \*.go -print)

.PHONY: mock
mock:
	mocker --prefix "" --dst mock/api_authorization.go --pkg mock api_authorization.go AuthorizationApi
	mocker --prefix "" --dst mock/api_cluster_registry.go --pkg mock api_cluster_registry.go ClusterRegistryApi
	mocker --prefix "" --dst mock/api_kafka_acl_management.go --pkg mock api_kafka_acl_management.go KafkaACLManagementApi
	mocker --prefix "" --dst mock/api_rbac_role_binding_crud.go --pkg mock api_rbac_role_binding_crud.go RBACRoleBindingCRUDApi
	mocker --prefix "" --dst mock/api_rbac_role_binding_summaries.go --pkg mock api_rbac_role_binding_summaries.go RBACRoleBindingSummariesApi
	mocker --prefix "" --dst mock/api_rbac_role_definitions.go --pkg mock api_rbac_role_definitions.go RBACRoleDefinitionsApi
	mocker --prefix "" --dst mock/api_tokens_and_authentication.go --pkg mock api_tokens_and_authentication.go TokensAndAuthenticationApi
	mocker --prefix "" --dst mock/api_audit_log_configuration.go --pkg mock api_audit_log_configuration.go AuditLogConfigurationApi

.PHONY: fmt
# Format entire codebase
fmt:
	@gofmt -e -s -l -w $(ALL_SRC)
