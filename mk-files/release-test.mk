.PHONY: verify-stag
verify-stag:
	OVERRIDE_S3_FOLDER=$(S3_STAG_FOLDER_NAME) make verify-archive-installer
	$(call dry-run,OVERRIDE_S3_FOLDER=$(S3_STAG_FOLDER_NAME) make smoke-tests)
	VERIFY_BIN_FOLDER=$(S3_STAG_PATH) make verify-binaries

.PHONY: verify-prod
verify-prod:
	OVERRIDE_S3_FOLDER="" make verify-archive-installer
	VERIFY_BIN_FOLDER=$(S3_BUCKET_PATH) make verify-binaries

.PHONY: verify-archive-installer
verify-archive-installer:
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION="" make test-installer
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION=v$(CLEAN_VERSION) make test-installer
	@echo "*** ARCHIVES VERIFICATION PASSED!!! ***"

# if ARCHIVES_VERSION is empty, latest folder will be tested
.PHONY: test-installer
test-installer:
	$(call dry-run,bash test-installer.sh $(ARCHIVES_VERSION))

# check that the expected binaries are present and have --acl public-read
.PHONY: verify-binaries
verify-binaries:
	$(eval DIR=$(shell mktemp -d))

	$(aws-authenticate) && \
	for os in linux alpine darwin windows; do \
		for arch in arm64 amd64; do \
			if [ "$${os}" = "windows" ] && [ "$${arch}" = "arm64" ] ; then \
				continue; \
			fi; \
			suffix=""; \
			if [ "$${os}" = "windows" ] ; then \
				suffix=".exe"; \
			fi; \
			FILE=$(VERIFY_BIN_FOLDER)/confluent-cli/binaries/$(CLEAN_VERSION)/confluent_$(CLEAN_VERSION)_$${os}_$${arch}$${suffix}; \
			echo "Checking binary: $${FILE}"; \
			$(call dry-run,aws s3 cp $$FILE $(DIR)) || { rm -rf $(DIR) && exit 1; }; \
		done; \
	done

	rm -rf $(DIR)	
	@echo "*** BINARIES VERIFICATION PASSED!!! ***"

# Test username/password login and SSO login in production
.PHONY: smoke-tests
smoke-tests:
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) bash install.sh && \
	export VAULT_ADDR=https://vault.cireops.gcp.internal.confluent.cloud && \
	vault login -method=oidc -path=okta && \
	password=$$(vault kv get -field password v1/devel/kv/cli/system-tests/test-user-password) && \
	email="cli-team+system-tests@confluent.io" && \
	echo -e "$${email}\n$${password}\n" | HOME=$$(mktemp -d) ./bin/confluent login && \
	go install github.com/confluentinc/cli-plugins/confluent-login-headless_sso@latest && \
	email="cli-team+system-tests+sso@confluent.io" && \
	HOME=$$(mktemp -d) ./bin/confluent login headless-sso --provider okta --email $${email} --password $${password}
