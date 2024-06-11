.PHONY: test-installer
test-installer:
	$(call dry-run,bash test-installer.sh)

# Test username/password login and SSO login in production
.PHONY: smoke-tests
smoke-tests:
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) bash install.sh && \
	export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud && \
	vault login -method=oidc -path=okta && \
	password=$$(vault kv get -field password v1/devel/kv/cli/system-tests/test-user-password) && \
	echo -e "cli-team+system-tests@confluent.io\n$${password}\n" | HOME=$$(mktemp -d) ./bin/confluent login && \
	go install github.com/confluentinc/cli-plugins/confluent-login-headless_sso@latest && \
	HOME=$$(mktemp -d) ./bin/confluent login headless-sso --provider okta --email cli-team+system-tests+sso@confluent.io --password $${password}
