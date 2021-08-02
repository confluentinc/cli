.PHONY: verify-stag
verify-stag:
	OVERRIDE_S3_FOLDER=$(S3_STAG_FOLDER_NAME) make verify-archive-installer
	VERIFY_BIN_FOLDER=$(S3_STAG_PATH) make verify-binary-files

.PHONY: verify-prod
verify-prod:
	OVERRIDE_S3_FOLDER="" make verify-archive-installer
	VERIFY_BIN_FOLDER=$(S3_BUCKET_PATH) make verify-binary-files

.PHONY: verify-archive-installer
verify-archive-installer:
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION="" make test-installer
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION=v$(CLEAN_VERSION) make test-installer
	@echo "*** ARCHIVES VERIFICATION PASSED!!! ***"

# if ARCHIVES_VERSION is empty, latest folder will be tested
.PHONY: test-installer
test-installer:
	@echo Running packaging/installer tests
	@bash test-installer.sh $(ARCHIVES_VERSION)

# check that the expected binaries are present and have --acl public-read
.PHONY: verify-binary
verify-binary:
	$(eval TEMP_DIR=$(shell mktemp -d))
	@$(caasenv-authenticate) && \
	binary="confluent"; \
	for os in linux darwin windows alpine; do \
		for arch in arm64 amd64 386; do \
			if [ "$${os}" != "darwin" ] && [ "$${arch}" = "arm64" ] ; then \
				continue; \
			fi ; \
			if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
				continue; \
			fi ; \
			if [ "$${os}" = "alpine" ] && [ "$${arch}" = "386" ] ; then \
				continue; \
			fi ; \
			suffix="" ; \
			if [ "$${os}" = "windows" ] ; then \
				suffix=".exe"; \
			fi ; \
			FILE=$(VERIFY_BIN_FOLDER)/$${binary}-cli/binaries/$(CLEAN_VERSION)/$${binary}_$(CLEAN_VERSION)_$${os}_$${arch}$${suffix}; \
			echo "Checking binary: $${FILE}"; \
			aws s3 cp $$FILE $(TEMP_DIR) || { rm -rf $(TEMP_DIR) && exit 1; }; \
		done; \
	done
	rm -rf $(TEMP_DIR)	
	@echo "*** BINARIES VERIFICATION PASSED!!! ***"

