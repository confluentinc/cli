ARCHIVE_TYPES=darwin_amd64.tar.gz darwin_arm64.tar.gz linux_amd64.tar.gz linux_arm64.tar.gz alpine_amd64.tar.gz alpine_arm64.tar.gz windows_amd64.zip

.PHONY: release
release: check-branch commit-release tag-release
	$(call print-boxed-message,"RELEASING TO STAGING FOLDER $(S3_STAG_PATH)")
	make release-to-stag
	$(call print-boxed-message,"PUBLISHING RELEASE NOTES TO S3 $(S3_BUCKET_PATH)")
	make publish-release-notes-to-s3
	$(call print-boxed-message,"RELEASING TO PROD FOLDER $(S3_BUCKET_PATH)")
	make release-to-prod
	$(call print-boxed-message,"PUBLISHING DOCS")
	@VERSION=$(VERSION) make publish-docs
	$(call print-boxed-message,"PUBLISHING NEW DOCKER HUB IMAGES")
	make publish-dockerhub

.PHONY: check-branch
check-branch:
	@if [ $(shell git rev-parse --abbrev-ref HEAD) != $(RELEASE_BRANCH) ] ; then \
		echo -n "WARNING: Current branch \"$(shell git rev-parse --abbrev-ref HEAD)\" is not the default release branch \"$(RELEASE_BRANCH)\"!  Do you want to proceed? (y/n): " ; \
		read line; if [ $$line != "y" ] && [ $$line != "Y" ]; then echo "Release cancelled."; exit 0; fi ; \
	fi

.PHONY: release-to-stag
release-to-stag:
	@make gorelease
	make goreleaser-patches
	make copy-stag-archives-to-latest
	$(call print-boxed-message,"VERIFYING STAGING RELEASE CONTENT")
	make verify-stag
	$(call print-boxed-message,"STAGING RELEASE COMPLETED AND VERIFIED!")

.PHONY: release-to-prod
release-to-prod:
	@$(aws-authenticate) && \
	$(call copy-stag-content-to-prod,archives,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,binaries,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,archives,latest)
	$(call print-boxed-message,"VERIFYING PROD RELEASE CONTENT")
	make verify-prod
	$(call print-boxed-message,"PROD RELEASE COMPLETED AND VERIFIED!")

define copy-stag-content-to-prod
	folder_path=confluent-cli/$1/$2; \
	echo "COPYING: $${folder_path}"; \
	aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_BUCKET_PATH)/$${folder_path} --recursive --acl public-read || exit 1
endef

# The glibc container doesn't need to publish to S3 so it doesn't need to $(caasenv-authenticate)
.PHONY: gorelease-linux-glibc
gorelease-linux-glibc:
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	VERSION=$(VERSION) GOEXPERIMENT=boringcrypto goreleaser release --clean -f .goreleaser-linux-glibc.yml

.PHONY: gorelease-linux-glibc-arm64
gorelease-linux-glibc-arm64:
ifneq (,$(findstring x86_64,$(shell uname -m)))
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	VERSION=$(VERSION) CGO_ENABLED=1 CC=aarch64-linux-gnu-gcc CXX=aarch64-linux-gnu-g++ GOEXPERIMENT=boringcrypto goreleaser release --clean -f .goreleaser-linux-glibc-arm64.yml
else
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	VERSION=$(VERSION) GOEXPERIMENT=boringcrypto goreleaser release --clean -f .goreleaser-linux-glibc-arm64.yml
endif

# This builds the Darwin, Windows and Linux binaries using goreleaser on the host computer. Goreleaser takes care of uploading the resulting binaries/archives/checksums to S3.
# Uploading linux glibc files because its goreleaser file has set release disabled
.PHONY: gorelease
gorelease:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	$(aws-authenticate) && \
	echo "BUILDING FOR DARWIN, WINDOWS, AND ALPINE LINUX" && \
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	VERSION=$(VERSION) GITHUB_TOKEN=$(token) S3FOLDER=$(S3_STAG_FOLDER_NAME)/confluent-cli GOEXPERIMENT=boringcrypto goreleaser release --clean --timeout 60m -f .goreleaser.yml; \
	rm -f CLIEVCodeSigningCertificate2.pfx && \
	echo "BUILDING FOR GLIBC LINUX" && \
	scripts/build_linux_glibc.sh && \
	aws s3 cp dist/confluent_$(VERSION_NO_V)_linux_amd64.tar.gz $(S3_STAG_PATH)/confluent-cli/archives/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_linux_amd64.tar.gz && \
	aws s3 cp dist/confluent_$(VERSION_NO_V)_linux_arm64.tar.gz $(S3_STAG_PATH)/confluent-cli/archives/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_linux_arm64.tar.gz && \
	$(aws-authenticate) && \
	aws s3 cp dist/confluent_linux_amd64_v1/confluent $(S3_STAG_PATH)/confluent-cli/binaries/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_linux_amd64 && \
	aws s3 cp dist/confluent_linux_arm64/confluent $(S3_STAG_PATH)/confluent-cli/binaries/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_linux_arm64 && \
	cat dist/confluent_$(VERSION_NO_V)_checksums_linux.txt >> dist/confluent_$(VERSION_NO_V)_checksums.txt && \
	cat dist/confluent_$(VERSION_NO_V)_checksums_linux_arm64.txt >> dist/confluent_$(VERSION_NO_V)_checksums.txt && \
	aws s3 cp dist/confluent_$(VERSION_NO_V)_checksums.txt $(S3_STAG_PATH)/confluent-cli/archives/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_checksums.txt && \
	aws s3 cp dist/confluent_$(VERSION_NO_V)_checksums.txt $(S3_STAG_PATH)/confluent-cli/binaries/$(VERSION_NO_V)/confluent_$(VERSION_NO_V)_checksums.txt && \
	echo "UPLOADING LINUX BUILDS TO GITHUB" && \
	make upload-linux-build-to-github
	

# Current goreleaser still has some shortcomings for the our use, and the target patches those issues
# As new goreleaser versions allow more customization, we may be able to reduce the work for this make target
.PHONY: goreleaser-patches
goreleaser-patches:
	make set-acls

# goreleaser does not yet support setting ACLs for cloud storage
# We have to set `public-read` manually by copying the file in place
# Dummy metadata is used as a hack because S3 does not allow copying files to the same place without any changes (--acl change doesn't count)
.PHONY: set-acls
set-acls:
	$(aws-authenticate) && \
	for file_type in binaries archives; do \
		folder_path=confluent-cli/$${file_type}/$(VERSION_NO_V); \
		echo "SETTING ACLS: $${folder_path}"; \
		aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_STAG_PATH)/$${folder_path} --acl public-read --metadata dummy=dummy --recursive || exit 1; \
	done

# Update latest archives folder for staging
# Also used by unrelease to fix latest archives folder so have to be careful about the version variable used
# e.g. may be using this to restore while cli repo VERSION value is dirty, hence CLEAN_VERSION variable is used
.PHONY: copy-stag-archives-to-latest
copy-stag-archives-to-latest:
	$(call copy-archives-files-to-latest,$(S3_STAG_PATH),$(S3_STAG_PATH))
	$(call copy-archives-checksums-to-latest,$(S3_STAG_PATH),$(S3_STAG_PATH))

# first argument: S3 folder of archives we want to copy from
# second argument: S3 folder destination for latest archives
define copy-archives-files-to-latest
	$(aws-authenticate); \
	archives_folder=$1/confluent-cli/archives/$(CLEAN_VERSION); \
	latest_folder=$2/confluent-cli/archives/latest; \
	for suffix in $(ARCHIVE_TYPES); do \
		aws s3 cp $${archives_folder}/confluent_$(CLEAN_VERSION)_$${suffix} $${latest_folder}/confluent_latest_$${suffix} --acl public-read; \
	done
endef

# Copy archives checksum file then rename the filenames in the checksum by replacing VERSION to "latest"
# Then publish the checksum file to S3 latest folder
# first argument: S3 folder of archives we want to copy from
# second argument: S3 folder destination for latest archives
define copy-archives-checksums-to-latest
	$(eval TEMP_DIR=$(shell mktemp -d))
	$(aws-authenticate); \
	version_checksums=confluent_$(CLEAN_VERSION)_checksums.txt; \
	latest_checksums=confluent_latest_checksums.txt; \
	cd $(TEMP_DIR) ; \
	aws s3 cp $1/confluent-cli/archives/$(CLEAN_VERSION)/$${version_checksums} ./ ; \
	cat $${version_checksums} | grep "$(CLEAN_VERSION)" | sed 's/$(CLEAN_VERSION)/latest/' > $${latest_checksums} ; \
	aws s3 cp $${latest_checksums} $2/confluent-cli/archives/latest/$${latest_checksums} --acl public-read
	rm -rf $(TEMP_DIR)
endef

.PHONY: download-licenses
download-licenses:
	go install github.com/google/go-licenses@v1.5.0 && \
	go-licenses save ./... --save_path legal/licenses --force || true

.PHONY: publish-installer
## Publish install scripts to S3. You MUST re-run this if/when you update any install script.
publish-installer:
	$(aws-authenticate) && \
	aws s3 cp install.sh $(S3_BUCKET_PATH)/confluent-cli/install.sh --acl public-read

.PHONY: upload-linux-build-to-github
## upload local copy of glibc linux build to github
upload-linux-build-to-github:
	gh release upload $(VERSION) dist/confluent_$(VERSION_NO_V)_linux_amd64.tar.gz && \
	gh release upload $(VERSION) dist/confluent_$(VERSION_NO_V)_linux_arm64.tar.gz && \
	mv dist/confluent_linux_amd64_v1/confluent dist/confluent_linux_amd64_v1/confluent_$(VERSION_NO_V)_linux_amd64 && \
	mv dist/confluent_linux_arm64/confluent dist/confluent_linux_arm64/confluent_$(VERSION_NO_V)_linux_arm64 && \
	gh release upload $(VERSION) dist/confluent_linux_amd64_v1/confluent_$(VERSION_NO_V)_linux_amd64 && \
	gh release upload $(VERSION) dist/confluent_linux_arm64/confluent_$(VERSION_NO_V)_linux_arm64 && \
	gh release upload $(VERSION) --clobber dist/confluent_$(VERSION_NO_V)_checksums.txt
