.PHONY: verify-stag
verify-stag:
	OVERRIDE_S3_FOLDER=$(S3_STAG_FOLDER_NAME) make verify-archive-installers
	VERIFY_BIN_FOLDER=$(S3_STAG_PATH) make verify-binary-files

.PHONY: verify-prod
verify-prod:
	OVERRIDE_S3_FOLDER="" make verify-archive-installers
	VERIFY_BIN_FOLDER=$(S3_BUCKET_PATH) make verify-binary-files

.PHONY: verify-archive-installers
verify-archive-installers:
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION="" make test-installers 
	OVERRIDE_S3_FOLDER=$(OVERRIDE_S3_FOLDER) ARCHIVES_VERSION=v$(CLEAN_VERSION) make test-installers 
	@echo "*** ARCHIVES VERIFICATION PASSED!!! ***"

# if ARCHIVES_VERSION is empty, latest folder will be tested
.PHONY: test-installers
test-installers:
	@echo Running packaging/installer tests
	@bash test-installers.sh $(ARCHIVES_VERSION)

# check that the expected binaries are present and have --acl public-read
.PHONY: verify-binary-files
verify-binary-files:
	$(eval TEMP_DIR=$(shell mktemp -d))
	@$(caasenv-authenticate) && \
	for binary in ccloud confluent; do \
		for os in linux darwin windows; do \
			for arch in amd64 386; do \
				if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
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
		done; \
	done
	rm -rf $(TEMP_DIR)	
	@echo "*** BINARIES VERIFICATION PASSED!!! ***"

