.PHONY: release
release: get-release-image commit-release tag-release
	@GO111MODULE=on make gorelease
	make set-acls
	make copy-archives-to-latest
	make rename-archives-checksums 
	git checkout go.sum
	@GO111MODULE=on VERSION=$(VERSION) make publish-docs
	git checkout go.sum

.PHONY: fakerelease
fakerelease: get-release-image commit-release tag-release
	@GO111MODULE=on make fakegorelease

S3_CCLOUD_FOLDER=ccloud-cli
S3_CONFLUENT_FOLDER=confluent-cli
ifeq (true, $(RELEASE_TEST))
S3_CCLOUD_FOLDER=$(S3_RELEASE_TEST_FOLDER)/ccloud-cli
S3_CONFLUENT_FOLDER=$(S3_RELEASE_TEST_FOLDER)/confluent-cli
endif

.PHONY: gorelease
gorelease:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	$(caasenv-authenticate) && \
	GO111MODULE=off go get -u github.com/inconshreveable/mousetrap && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_CCLOUD_FOLDER) goreleaser release --rm-dist -f .goreleaser-ccloud.yml && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(S3_CONFLUENT_FOLDER) goreleaser release --rm-dist -f .goreleaser-confluent.yml

# goreleaser does not yet support setting ACLs for S3 so we have set `public-read` manually by copy the file in place
# dummy metadata is used as a hack because S3 does not allow copying files to the same place without any changes (--acl change not included)
.PHONY: set-acls
set-acls:
	$(caasenv-authenticate) && \
	aws s3 cp $(S3_BUCKET_PATH)/ccloud-cli/binaries/$(VERSION_NO_V) $(S3_BUCKET_PATH)/ccloud-cli/binaries/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive && \
	aws s3 cp $(S3_BUCKET_PATH)/confluent-cli/binaries/$(VERSION_NO_V) $(S3_BUCKET_PATH)/confluent-cli/binaries/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive && \
	aws s3 cp $(S3_BUCKET_PATH)/ccloud-cli/archives/$(VERSION_NO_V) $(S3_BUCKET_PATH)/ccloud-cli/archives/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive && \
	aws s3 cp $(S3_BUCKET_PATH)/confluent-cli/archives/$(VERSION_NO_V) $(S3_BUCKET_PATH)/confluent-cli/archives/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive

.PHONY: copy-archives-to-latest
copy-archives-to-latest:
	$(eval TEMP_DIR=$(shell mktemp -d))
	$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		aws s3 cp $(S3_BUCKET_PATH)/$${binary}-cli/archives/$(CLEAN_VERSION) $(TEMP_DIR)/$${binary}-cli --recursive ; \
		cd $(TEMP_DIR)/$${binary}-cli ; \
		for fname in $${binary}_v$(CLEAN_VERSION)_*; do \
			newname=`echo "$$fname" | sed 's/_v$(CLEAN_VERSION)/_latest/g'`; \
			mv $$fname $$newname; \
		done ; \
		rm *checksums.txt; \
		$(SHASUM) $${binary}_latest_* > $${binary}_latest_checksums.txt ; \
		aws s3 cp ./ $(S3_BUCKET_PATH)/$${binary}-cli/archives/latest --acl public-read --recursive ; \
	done
	rm -rf $(TEMP_DIR)

# goreleaser uploads the checksum for archives as ccloud_1.19.0_checksums.txt but the installer script expects version with 'v', i.e. ccloud_v1.19.0_checksums.txt
# chose to not change install script because older versions uses the no-v format
# if we update the script to accept both checksums name format, this target would no longer be needed
.PHONY: rename-archives-checksums
rename-archives-checksums:
	$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		aws s3 mv $(S3_BUCKET_PATH)/$${binary}-cli/archives/$(VERSION_NO_V)/$${binary}_$(VERSION_NO_V)_checksums.txt $(S3_BUCKET_PATH)/$${binary}-cli/archives/$(VERSION_NO_V)/$${binary}_$(VERSION)_checksums.txt --acl public-read; \
	done


.PHONY: fakegorelease
fakegorelease:
	@GO111MODULE=off go get -u github.com/inconshreveable/mousetrap # dep from cobra -- incompatible with go mod
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-ccloud-fake.yml
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-confluent-fake.yml

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
