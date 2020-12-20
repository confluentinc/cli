ARCHIVE_TYPES=darwin_amd64.tar.gz linux_amd64.tar.gz linux_386.tar.gz windows_amd64.zip windows_386.zip alpine_amd64.tar.gz

.PHONY: release
release: get-release-image commit-release tag-release
	$(call print-boxed-message,"RELEASING TO STAGING FOLDER $(S3_STAG_PATH)")
	make release-to-stag
	$(call print-boxed-message,"RELEASING TO PROD FOLDER $(S3_BUCKET_PATH)")
	make release-to-prod
	$(call print-boxed-message,"PUBLISHING DOCS")
	@GO111MODULE=on VERSION=$(VERSION) make publish-docs
	git checkout go.sum
	$(call print-boxed-message,"PUBLISHING NEW DOCKER HUB IMAGES")
	make publish-dockerhub

.PHONY: release-to-stag
release-to-stag:
	@GO111MODULE=on make gorelease
	git checkout go.sum
	make goreleaser-patches
	make copy-stag-archives-to-latest
	$(call print-boxed-message,"VERIFYING STAGING RELEASE CONTENT")
	make verify-stag
	$(call print-boxed-message,"STAGING RELEASE COMPLETED AND VERIFIED!")

.PHONY: release-to-prod
release-to-prod:
	@$(caasenv-authenticate) && \
	$(call copy-stag-content-to-prod,archives,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,binaries,$(CLEAN_VERSION)); \
	$(call copy-stag-content-to-prod,archives,latest)
	$(call print-boxed-message,"VERIFYING PROD RELEASE CONTENT")
	make verify-prod
	$(call print-boxed-message,"PROD RELEASE COMPLETED AND VERIFIED!")

define copy-stag-content-to-prod
	for binary in ccloud confluent; do \
		folder_path=$${binary}-cli/$1/$2; \
		echo "COPYING: $${folder_path}"; \
		aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_BUCKET_PATH)/$${folder_path} --recursive --acl public-read || exit 1; \
	done
endef

# The Alpine container doesn't need to publish to S3 so it doesn't need to $(caasenv-authenticate)
.PHONY: gorelease-alpine
gorelease-alpine:
	GO111MODULE=off go get -u github.com/inconshreveable/mousetrap && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_STAG_FOLDER_NAME)/ccloud-cli goreleaser release --rm-dist -f .goreleaser-ccloud-alpine.yml && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_STAG_FOLDER_NAME)/confluent-cli goreleaser release --rm-dist -f .goreleaser-confluent-alpine.yml

# This builds the Darwin, Linux (non-Alpine), and Windows binaries using goreleaser on the host computer.  goreleaser takes care of uploading the resulting binaries/archives/checksums to S3.  However, we then have to separately build the Alpine binaries/archives in a Docker container (since we need to use an OS which has the Alpine C runtimes instead of the C runtimes on macOS).  We then also have to manually upload the Alpine build artifacts to S3 since the goreleaser inside the Docker container doesn't have the S3 credentials from the host.
.PHONY: gorelease
gorelease:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	$(caasenv-authenticate) && \
	GO111MODULE=off go get -u github.com/inconshreveable/mousetrap && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_STAG_FOLDER_NAME)/ccloud-cli goreleaser release --rm-dist -f .goreleaser-ccloud.yml && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_STAG_FOLDER_NAME)/confluent-cli goreleaser release --rm-dist -f .goreleaser-confluent.yml && \
	./build_alpine.sh && \
	for binary in ccloud confluent; do \
		aws s3 cp dist/$${binary}/$${binary}_$(VERSION)_alpine_amd64.tar.gz $(S3_STAG_PATH)/$${binary}-cli/archives/$(VERSION_NO_V)/$${binary}_$(VERSION)_alpine_amd64.tar.gz; \
		aws s3 cp dist/$${binary}/$${binary}_alpine_amd64/$${binary} $(S3_STAG_PATH)/$${binary}-cli/binaries/$(VERSION_NO_V)/$${binary}_$(VERSION_NO_V)_alpine_amd64; \
		cat dist/$${binary}/$${binary}_$(VERSION_NO_V)_checksums_alpine.txt >> dist/$${binary}/$${binary}_$(VERSION_NO_V)_checksums.txt; \
	done

# Current goreleaser still has some shortcomings for the our use, and the target patches those issues
# As new goreleaser versions allow more customization, we may be able to reduce the work for this make target
.PHONY: goreleaser-patches
goreleaser-patches:
	make set-acls
	make rename-archives-checksums

# goreleaser does not yet support setting ACLs for cloud storage
# We have to set `public-read` manually by copying the file in place
# Dummy metadata is used as a hack because S3 does not allow copying files to the same place without any changes (--acl change doesn't count)
.PHONY: set-acls
set-acls:
	$(caasenv-authenticate) && \
	for binary in ccloud confluent; do \
		for file_type in binaries archives; do \
			folder_path=$${binary}-cli/$${file_type}/$(VERSION_NO_V); \
			echo "SETTING ACLS: $${folder_path}"; \
			aws s3 cp $(S3_STAG_PATH)/$${folder_path} $(S3_STAG_PATH)/$${folder_path} --acl public-read --metadata dummy=dummy --recursive || exit 1; \
		done; \
	done

# goreleaser uploads the checksum for archives as ccloud_1.19.0_checksums.txt but the installer script expects version with 'v', i.e. ccloud_v1.19.0_checksums.txt
# Chose not to change install script to expect no-v because older versions use the format with 'v'.
# Also, we first re-upload the checksums file because we concatenate the Alpine checksums to the checksums file after goreleaser has already published it (without the Alpine checksums) to S3
.PHONY: rename-archives-checksums
rename-archives-checksums:
	$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		folder=$(S3_STAG_PATH)/$${binary}-cli/archives/$(CLEAN_VERSION); \
		aws s3 cp dist/$${binary}/$${binary}_$(VERSION_NO_V)_checksums.txt $${folder}/$${binary}_$(CLEAN_VERSION)_checksums.txt;\
		aws s3 mv $${folder}/$${binary}_$(CLEAN_VERSION)_checksums.txt $${folder}/$${binary}_v$(CLEAN_VERSION)_checksums.txt --acl public-read; \
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
	$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		archives_folder=$1/$${binary}-cli/archives/$(CLEAN_VERSION); \
		latest_folder=$2/$${binary}-cli/archives/latest; \
		for suffix in $(ARCHIVE_TYPES); do \
			aws s3 cp $${archives_folder}/$${binary}_v$(CLEAN_VERSION)_$${suffix} $${latest_folder}/$${binary}_latest_$${suffix} --acl public-read; \
		done ; \
	done
endef

# Copy archives checksum file then rename the filenames in the checksum by replacing VERSION to "latest"
# Then publish the checksum file to S3 latest folder
# first argument: S3 folder of archives we want to copy from
# second argument: S3 folder destination for latest archives
define copy-archives-checksums-to-latest
	$(eval TEMP_DIR=$(shell mktemp -d))
	$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		version_checksums=$${binary}_v$(CLEAN_VERSION)_checksums.txt; \
		latest_checksums=$${binary}_latest_checksums.txt; \
		cd $(TEMP_DIR) ; \
		aws s3 cp $1/$${binary}-cli/archives/$(CLEAN_VERSION)/$${version_checksums} ./ ; \
		cat $${version_checksums} | grep "v$(CLEAN_VERSION)" | sed 's/v$(CLEAN_VERSION)/latest/' > $${latest_checksums} ; \
		aws s3 cp $${latest_checksums} $2/$${binary}-cli/archives/latest/$${latest_checksums} --acl public-read ; \
	done
	rm -rf $(TEMP_DIR)
endef

.PHONY: download-licenses
download-licenses:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	@# we'd like to use golicense -plain but the exit code is always 0 then so CI won't actually fail on illegal licenses
	@ echo Downloading third-party licenses for $(LICENSE_BIN) binary ; \
	GITHUB_TOKEN=$(token) golicense .golicense.hcl $(LICENSE_BIN_PATH) | GITHUB_TOKEN=$(token) go run cmd/golicense-downloader/main.go -F .golicense-downloader.json -l legal/licenses -n legal/notices ; \
	[ -z "$$(ls -A legal/licenses)" ] && { echo "ERROR: licenses folder not populated" && exit 1; }; \
	[ -z "$$(ls -A legal/notices)" ] && { echo "ERROR: notices folder not populated" && exit 1; }; \
	echo Successfully downloaded licenses

.PHONY: publish-installers
## Publish install scripts to S3. You MUST re-run this if/when you update any install script.
publish-installers:
	$(caasenv-authenticate) && \
	aws s3 cp install-ccloud.sh $(S3_BUCKET_PATH)/ccloud-cli/install.sh --acl public-read && \
	aws s3 cp install-confluent.sh $(S3_BUCKET_PATH)/confluent-cli/install.sh --acl public-read

.PHONY: docs
docs: clean-docs
	@GO111MODULE=on go run -ldflags '-X main.cliName=ccloud' cmd/docs/main.go
	@GO111MODULE=on go run -ldflags '-X main.cliName=confluent' cmd/docs/main.go

.PHONY: publish-docs
publish-docs: docs
	@tmp=$$(mktemp -d); \
	git clone git@github.com:confluentinc/docs.git $$tmp; \
	echo -n "Publish ccloud docs? (y/n) "; read line; \
	if [ $$line = "y" ] || [ $$line = "Y" ]; then make publish-docs-internal REPO_DIR=$$tmp CLI_NAME=ccloud; fi; \
	echo -n "Publish confluent docs? (y/n) "; read line; \
	if [ $$line = "y" ] || [ $$line = "Y" ]; then make publish-docs-internal REPO_DIR=$$tmp CLI_NAME=confluent; fi; \
	rm -rf $$tmp

.PHONY: publish-docs-internal
publish-docs-internal:
ifeq ($(CLI_NAME), ccloud)
	$(eval DOCS_DIR := ccloud-cli/command-reference)
else
	$(eval DOCS_DIR := confluent-cli/command-reference)
endif

	@cd $(REPO_DIR); \
	git checkout -b $(CLI_NAME)-cli-$(VERSION) origin/$(DOCS_BRANCH) || exit 1; \
	rm -rf $(DOCS_DIR); \
	cp -R $(GOPATH)/src/github.com/confluentinc/cli/docs/$(CLI_NAME) $(DOCS_DIR); \
	[ ! -f "$(DOCS_DIR)/kafka/topic/ccloud_kafka_topic_consume.rst" ] || sed -i '' 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<uuid>"/' $(DOCS_DIR)/kafka/topic/ccloud_kafka_topic_consume.rst || exit 1; \
	git add . || exit 1; \
	git diff --cached --exit-code > /dev/null && echo "nothing to update for docs" && exit 0; \
	git commit -m "chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1; \
	git push origin $(CLI_NAME)-cli-$(VERSION) || exit 1; \
	hub pull-request -b $(DOCS_BRANCH) -m "chore: update $(CLI_NAME) CLI docs for $(VERSION)" || exit 1

.PHONY: clean-docs
clean-docs:
	@rm -rf docs/

