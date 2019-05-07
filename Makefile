SHELL           := /bin/bash
ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin
MASTER_BRANCH   ?= master
RELEASE_BRANCH  ?= master

DOCS_BRANCH     ?= master

include ./semver.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u)
HOSTNAME := $(shell id -u -n)@$(shell hostname)

.PHONY: clean
clean:
	rm -rf $(shell pwd)/dist
	rm -f internal/cmd/local/bindata.go
	rm -f mock/local/shell_runner_mock.go

.PHONY: generate
generate: generate-go bindata mocks

.PHONY: generate-go
generate-go:
	@go generate ./...

.PHONY: deps
deps:
	@GO111MODULE=on go get github.com/goreleaser/goreleaser@v0.106.0
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0
	@GO111MODULE=on go get github.com/mitchellh/golicense@v0.1.1
	@GO111MODULE=on go get github.com/golang/mock/mockgen@v1.2.0
	@GO111MODULE=on go get github.com/kevinburke/go-bindata/...@v3.13.0

build: build-go

ifeq ($(shell uname),Darwin)
GORELEASER_SUFFIX ?= -mac.yml
SHASUM ?= gsha256sum
else ifneq (,$(findstring NT,$(shell uname)))
GORELEASER_SUFFIX ?= -windows.yml
# TODO: I highly doubt this works. Completely untested. The output format is likely very different than expected.
SHASUM ?= CertUtil SHA256 -hashfile
else
GORELEASER_SUFFIX ?= -linux.yml
SHASUM ?= sha256sum
endif

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: build-go
build-go:
	make build-ccloud
	make build-confluent

.PHONY: build-ccloud
build-ccloud:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-ccloud$(GORELEASER_SUFFIX)

.PHONY: build-confluent
build-confluent:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-confluent$(GORELEASER_SUFFIX)

.PHONY: bindata
bindata: internal/cmd/local/bindata.go

internal/cmd/local/bindata.go:
	@go-bindata -pkg local -o internal/cmd/local/bindata.go cp_cli/

.PHONY: release
release: get-release-image commit-release tag-release
	make gorelease
	make publish
	make publish-docs

.PHONY: gorelease
gorelease:
	@GO111MODULE=off go get -u github.com/inconshreveable/mousetrap # dep from cobra -- incompatible with go mod
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-ccloud.yml
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist -f .goreleaser-confluent.yml

.PHONY: download-licenses
download-licenses:
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	@for binary in ccloud confluent; do \
		echo Downloading third-party licenses for $${binary} binary ; \
		GITHUB_TOKEN=$(token) golicense .golicense.hcl ./dist/$${binary}/$(shell go env GOOS)_$(shell go env GOARCH)/$${binary} | go run cmd/license-downloader/main.go -l legal/$${binary}/licenses -n legal/$${binary}/notices ; \
		[ -z "$$(ls -A legal/$${binary}/licenses)" ] && rmdir legal/$${binary}/licenses ; \
		[ -z "$$(ls -A legal/$${binary}/notices)" ] && rmdir legal/$${binary}/notices ; \
	done

.PHONY: dist
dist: download-licenses
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	@for binary in ccloud confluent; do \
		for os in $$(find dist/$${binary} -type d -mindepth 1 -maxdepth 1 | awk -F'/' '{ print $$3 }' | awk -F'_' '{ print $$1 }'); do \
			for arch in $$(find dist/$${binary} -type d -mindepth 1 -maxdepth 1 -iname $${os}_* | awk -F'/' '{ print $$3 }' | awk -F'_' '{ print $$2 }'); do \
				if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
					continue ; \
				fi; \
				[ "$${os}" = "windows" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
				rm -rf /tmp/$${binary} && mkdir /tmp/$${binary} ; \
				cp LICENSE /tmp/${binary} && cp -r legal/$${binary} /tmp/$${binary}/legal ; \
				cp dist/$${binary}/$${os}_$${arch}/$${binexe} /tmp/$${binary} ; \
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
publish:
	@for binary in ccloud confluent; do \
		aws s3 cp dist/$(NAME)/ s3://confluent.cloud/$(NAME)-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_latest_*" --acl public-read ; \
		aws s3 cp dist/$(NAME)/ s3://confluent.cloud/$(NAME)-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --include "*_checksums.txt" --exclude "*_$(VERSION)_*" --acl public-read ; \
	done

.PHONY: publish-installers
## Publish install scripts to S3. You MUST re-run this if/when you update any install script.
publish-installers:
	aws s3 cp install-ccloud.sh s3://confluent.cloud/ccloud-cli/install.sh --acl public-read
	aws s3 cp install-confluent.sh s3://confluent.cloud/confluent-cli/install.sh --acl public-read

.PHONY: docs
docs:
#   TODO: we can't enable auto-docs generation for confluent until we migrate go-basher commands into cobra
#	@GO111MODULE=on go run -ldflags '-X main.cliName=confluent' cmd/docs/main.go
	@GO111MODULE=on go run -ldflags '-X main.cliName=ccloud' cmd/docs/main.go

.PHONY: publish-docs
publish-docs: docs
	@TMP_DIR=$$(mktemp -d)/docs || exit 1; \
		git clone git@github.com:confluentinc/docs.git $${TMP_DIR}; \
		cd $${TMP_DIR} || exit 1; \
		git checkout -b cli-$(VERSION) $(DOCS_BRANCH) || exit 1; \
		cd - || exit 1; \
		make publish-docs-internal BASE_DIR=$${TMP_DIR} CLI_NAME=ccloud || exit 1; \
		cd $${TMP_DIR} || exit 1; \
		sed -i 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<uuid>"/' cloud/cli/command-reference/ccloud_kafka_topic_consume.rst || exit 1; \
		git add . || exit 1; \
		git diff --cached --exit-code >/dev/null && echo "nothing to update for docs" && exit 0; \
		git commit -m "chore: updating CLI docs for $(VERSION)" || exit 1; \
		git push origin cli-$(VERSION) || exit 1; \
		hub pull-request -b $(DOCS_BRANCH) -m "chore: updating CLI docs for $(VERSION)" || exit 1; \
		cd - || exit 1; \
		rm -rf $${TMP_DIR}
#   TODO: we can't enable auto-docs generation for confluent until we migrate go-basher commands into cobra
#	    make publish-docs-internal BASE_DIR=$${TMP_DIR} CLI_NAME=confluent || exit 1; \

.PHONY: publish-docs-internal
publish-docs-internal:
ifndef BASE_DIR
	$(error BASE_DIR is not set)
endif
ifeq (ccloud,$(CLI_NAME))
	$(eval DOCS_DIR := cloud/cli/command-reference)
else ifeq (confluent,$(CLI_NAME))
	$(eval DOCS_DIR := cli/command-reference)
else
	$(error CLI_NAME is not set correctly - must be one of "confluent" or "ccloud")
endif
	rm $(BASE_DIR)/$(DOCS_DIR)/*.rst
	cp $(GOPATH)/src/github.com/confluentinc/cli/docs/$(CLI_NAME)/*.rst $(BASE_DIR)/$(DOCS_DIR)

.PHONY: clean-docs
clean-docs:
	rm docs/*/*.rst

.PHONY: fmt
fmt:
	@gofmt -e -s -l -w $(ALL_SRC)

.PHONY: release-ci
release-ci:
ifeq ($(SEMAPHORE_GIT_BRANCH),master)
	make release
else
	true
endif

cmd/lint/en_US.aff:
	@curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.aff?format=TEXT" | base64 -D > $@

cmd/lint/en_US.dic:
	@curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.dic?format=TEXT" | base64 -D > $@

.PHONY: lint-cli
lint-cli: cmd/lint/en_US.aff cmd/lint/en_US.dic
	GO111MODULE=on go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)

.PHONY: lint-go
lint-go:
	@GO111MODULE=on golangci-lint run

.PHONY: lint
lint: lint-go lint-installers

.PHONY: lint-installers
## Lints the CLI installation scripts
lint-installers:
	@diff install-c* | grep -v -E "^---|^[0-9c0-9]|PROJECT_NAME|BINARY" && echo "diff between install scripts" && exit 1 || exit 0

.PHONY: lint-licenses
## Scan and validate third-party dependeny licenses
lint-licenses: build
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	@for binary in ccloud confluent; do \
		echo Licenses for $${binary} binary ; \
		GITHUB_TOKEN=$(token) golicense .golicense.hcl ./dist/$${binary}/$(shell go env GOOS)_$(shell go env GOARCH)/$${binary} ; \
	done

.PHONY: coverage
coverage:
      ifdef CI
	@echo "" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor); do \
	  GO111MODULE=on go test -v -race -coverprofile=profile.out -covermode=atomic $$d || exit 2; \
	  if [ -f profile.out ]; then \
	    cat profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
      else
	@GO111MODULE=on go test -race -cover $(TEST_ARGS) $$(go list ./... | grep -v vendor)
      endif

.PHONY: mocks
mocks: mock/local/shell_runner_mock.go

mock/local/shell_runner_mock.go:
	mockgen -source internal/cmd/local/shell_runner.go -destination mock/local/shell_runner_mock.go ShellRunner

.PHONY: test
test: lint coverage
