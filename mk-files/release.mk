.PHONY: release
release: get-release-image commit-release tag-release
	@GO111MODULE=on make gorelease
	git checkout go.sum
	@GO111MODULE=on VERSION=$(VERSION) make publish
	@GO111MODULE=on VERSION=$(VERSION) make publish-docs
	git checkout go.sum

.PHONY: fakerelease
fakerelease: get-release-image commit-release tag-release
	@GO111MODULE=on make fakegorelease

GORELEASE_S3_CCLOUD_FOLDER=ccloud-cli/binaries
GORELEASE_S3_CONFLUENT_FOLDER=confluent-cli/binaries
ifeq (true, $(RELEASE_TEST))
GORELEASE_S3_CCLOUD_FOLDER=$(S3_RELEASE_TEST_FOLDER)/ccloud-cli/binaries
GORELEASE_S3_CONFLUENT_FOLDER=$(S3_RELEASE_TEST_FOLDER)/confluent-cli/binaries
endif

.PHONY: gorelease
gorelease:
	$(caasenv-authenticate) && \
	GO111MODULE=off go get -u github.com/inconshreveable/mousetrap && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(GORELEASE_S3_CCLOUD_FOLDER) goreleaser release --rm-dist -f .goreleaser-ccloud.yml && \
	GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" S3FOLDER=$(GORELEASE_S3_CONFLUENT_FOLDER) goreleaser release --rm-dist -f .goreleaser-confluent.yml && \
	aws s3 cp $(S3_BUCKET_PATH)/ccloud-cli/binaries/$(VERSION_NO_V) $(S3_BUCKET_PATH)/ccloud-cli/binaries/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive && \
	aws s3 cp $(S3_BUCKET_PATH)/confluent-cli/binaries/$(VERSION_NO_V) $(S3_BUCKET_PATH)/confluent-cli/binaries/$(VERSION_NO_V) --acl public-read --metadata dummy=dummy --recursive

.PHONY: fakegorelease
fakegorelease:
	@GO111MODULE=off go get -u github.com/inconshreveable/mousetrap # dep from cobra -- incompatible with go mod
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-ccloud-fake.yml
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-confluent-fake.yml

.PHONY: sign
sign:
	@GO111MODULE=on gon gon_ccloud.hcl
	@GO111MODULE=on gon gon_confluent.hcl
	rm dist/ccloud/ccloud_darwin_amd64/ccloud_signed.zip || true
	rm dist/confluent/confluent_darwin_amd64/confluent_signed.zip || true

.PHONY: download-licenses
download-licenses:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	@# we'd like to use golicense -plain but the exit code is always 0 then so CI won't actually fail on illegal licenses
	@for binary in ccloud confluent; do \
		echo Downloading third-party licenses for $${binary} binary ; \
		GITHUB_TOKEN=$(token) golicense .golicense.hcl ./dist/$${binary}/$${binary}_$(shell go env GOOS)_$(shell go env GOARCH)/$${binary} | GITHUB_TOKEN=$(token) go run cmd/golicense-downloader/main.go -F .golicense-downloader.json -l legal/$${binary}/licenses -n legal/$${binary}/notices ; \
		[ -z "$$(ls -A legal/$${binary}/licenses)" ] && rmdir legal/$${binary}/licenses ; \
		[ -z "$$(ls -A legal/$${binary}/notices)" ] && rmdir legal/$${binary}/notices ; \
		echo ; \
	done

.PHONY: dist
dist: download-licenses
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	@for binary in ccloud confluent; do \
		for os in $$(find dist/$${binary} -mindepth 1 -maxdepth 1 -type d | awk -F'/' '{ print $$3 }' | awk -F'_' '{ print $$2 }'); do \
			for arch in $$(find dist/$${binary} -mindepth 1 -maxdepth 1 -iname $${binary}_$${os}_* -type d | awk -F'/' '{ print $$3 }' | awk -F'_' '{ print $$3 }'); do \
				if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
					continue ; \
				fi; \
				[ "$${os}" = "windows" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
				rm -rf /tmp/$${binary} && mkdir /tmp/$${binary} ; \
				cp LICENSE /tmp/$${binary} && cp -r legal/$${binary} /tmp/$${binary}/legal ; \
				cp dist/$${binary}/$${binary}_$${os}_$${arch}/$${binexe} /tmp/$${binary} ; \
				suffix="" ; \
				if [ "$${os}" = "windows" ] ; then \
					suffix=zip ; \
					cd /tmp >/dev/null && zip -qr $${binary}.$${suffix} $${binary} && cd - >/dev/null ; \
					mv /tmp/$${binary}.$${suffix} dist/$${binary}/$${binary}_$(VERSION)_$${os}_$${arch}.$${suffix}; \
				else \
					suffix=tar.gz ; \
					tar -czf dist/$${binary}/$${binary}_$(VERSION)_$${os}_$${arch}.$${suffix} -C /tmp $${binary} ; \
				fi ; \
				cp dist/$${binary}/$${binary}_$(VERSION)_$${os}_$${arch}.$${suffix} dist/$${binary}/$${binary}_latest_$${os}_$${arch}.$${suffix} ; \
			done ; \
		done ; \
		cd dist/$${binary}/ ; \
		  $(SHASUM) $${binary}_$(VERSION)_* > $${binary}_$(VERSION)_checksums.txt ; \
		  $(SHASUM) $${binary}_latest_* > $${binary}_latest_checksums.txt ; \
		  cd ../.. ; \
	done

.PHONY: publish
## Note: gorelease target publishes unsigned binaries to the binaries folder in the bucket, we have to overwrite them here after signing
publish: sign dist
	@$(caasenv-authenticate); \
	for binary in ccloud confluent; do \
		aws s3 cp dist/$${binary}/$${binary}_darwin_amd64/$${binary} $(S3_BUCKET_PATH)/$${binary}-cli/binaries/$(VERSION:v%=%)/$${binary}_$(VERSION:v%=%)_darwin_amd64 --acl public-read ; \
		aws s3 cp dist/$${binary}/ $(S3_BUCKET_PATH)/$${binary}-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_latest_*" --acl public-read ; \
		aws s3 cp dist/$${binary}/ $(S3_BUCKET_PATH)/$${binary}-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_$(VERSION)_*" --acl public-read ; \
	done

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
