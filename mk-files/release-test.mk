VERIFY_RELEASE_TARGET ?= $(S3_BUCKET_PATH)
VERIFY_ARCHIVES_FOLDER_OVERRIDE ?=

.PHONY: test-installers
test-installers:
	@echo Running packaging/installer tests
	@ # if ${ARCHIVES_VERSION_TO_TEST} archives latest folder will be tested
	@bash test-installers.sh ${ARCHIVES_VERSION_TO_TEST}

.PHONY: verify-staging
verify-staging:
	VERIFY_ARCHIVES_FOLDER_TARGET=$(S3_STAG_FOLDER_NAME) make verify-archive-installers
	VERIFY_RELEASE_TARGET=$(S3_STAG_PATH) make verify-binary-files

.PHONY: verify-release
verify-release:
	make verify-archive-installers
	make verify-binary-files

.PHONY: verify-archive-installers
verify-archive-installers:
	make test-installers OVERRIDE_S3_FOLDER=$(VERIFY_ARCHIVES_FOLDER_TARGET)
	make test-installers ARCHIVES_VERSION_TO_TEST=v$(CLEAN_VERSION) OVERRIDE_S3_FOLDER=$(VERIFY_ARCHIVES_FOLDER_TARGET)
	@echo "ARCHIVES VERIFICATION PASSED!!!"

# check that the expected binaries are present and have --acl public-read
.PHONY: verify-binary-files
verify-binary-files:
	$(eval TEMP_DIR=$(shell mktemp -d))
	echo $(TEMP_DIR)
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
				FILE=$(VERIFY_RELEASE_TARGET)/$${binary}-cli/binaries/$(CLEAN_VERSION)/$${binary}_$(CLEAN_VERSION)_$${os}_$${arch}$${suffix}; \
				echo "Checking binary: $${FILE}"; \
				aws s3 cp $$FILE $(TEMP_DIR) || { rm -rf $(TEMP_DIR) && exit 1; }; \
			done; \
		done; \
	done
	rm -rf $(TEMP_DIR)	
	@echo "BINARY VERIFICATION PASSED!!!"
