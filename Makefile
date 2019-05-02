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

.PHONY: deps
deps:
	@GO111MODULE=on go get github.com/goreleaser/goreleaser@v0.106.0
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.16.0
	@GO111MODULE=on go get github.com/mitchellh/golicense@v0.1.1
	@GO111MODULE=on go get github.com/golang/mock/mockgen@v1.2.0
	@GO111MODULE=on go get github.com/kevinburke/go-bindata@v3.13.0

build: build-go

ifeq ($(shell uname),Darwin)
GORELEASER_SUFFIX ?= -mac.yml
else ifneq (,$(findstring NT,$(shell uname)))
GORELEASER_SUFFIX ?= -windows.yml
else
GORELEASER_SUFFIX ?= -linux.yml
endif

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: build-go
build-go: internal/cmd/local/bindata.go
	make build-ccloud
	make build-confluent

.PHONY: build-ccloud
build-ccloud:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-ccloud$(GORELEASER_SUFFIX)

.PHONY: build-confluent
build-confluent:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f .goreleaser-confluent$(GORELEASER_SUFFIX)

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

.PHONY: dist-ccloud
dist-ccloud:
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	for os in darwin linux windows; do \
		for arch in amd64 386; do \
			if [ "$${os}" = "darwin" ] && [ "$${arch}" = "386" ] ; then \
				continue ; \
			fi; \
			cp LICENSE dist/ccloud/$${os}_$${arch}/ ; \
			cp INSTALL.md dist/ccloud/$${os}_$${arch}/ ; \
			cd dist/ccloud/$${os}_$${arch}/ ; \
			mkdir tmp ; mv LICENSE INSTALL.md ccloud* tmp/ ; mv tmp ccloud ; \
			suffix="" ; \
			if [ "$${os}" = "windows" ] ; then \
				suffix=zip ; \
				zip -qr ../ccloud_$(VERSION)_$${os}_$${arch}.$${suffix} ccloud ; \
			else \
				suffix=tar.gz ; \
				tar -czf ../ccloud_$(VERSION)_$${os}_$${arch}.$${suffix} ccloud ; \
			fi ; \
			cd ../../../ ; \
			cp dist/ccloud/ccloud_$(VERSION)_$${os}_$${arch}.$${suffix} dist/ccloud/ccloud_latest_$${os}_$${arch}.$${suffix} ; \
		done ; \
	done

.PHONY: publish-ccloud
publish: dist-ccloud
	aws s3 cp dist/ccloud/ s3://confluent.cloud/ccloud-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_latest_*" --acl public-read
	aws s3 cp dist/ccloud/ s3://confluent.cloud/ccloud-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_$(VERSION)_*" --acl public-read

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
		sed -i '' 's/default "confluent_cli_consumer_[^"]*"/default "confluent_cli_consumer_<uuid>"/' cloud/cli/command-reference/ccloud_kafka_topic_consume.rst || exit 1; \
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
lint: lint-go

.PHONY: lint-licenses
## Scan and validate third-party dependeny licenses
lint-licenses: build
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	@echo Licenses for ccloud binary
	@GITHUB_TOKEN=$(token) golicense .golicense.hcl ./dist/ccloud/$(shell go env GOOS)_$(shell go env GOARCH)/ccloud
	@echo Licenses for confluent binary
	@GITHUB_TOKEN=$(token) golicense .golicense.hcl ./dist/confluent/$(shell go env GOOS)_$(shell go env GOARCH)/confluent

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
test: internal/cmd/local/bindata.go mocks lint coverage
