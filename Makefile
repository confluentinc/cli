SHELL           := /bin/bash
ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin
MASTER_BRANCH   ?= master
RELEASE_BRANCH  ?= master

DOCS_BRANCH     ?= 6.0.0-post

include ./mk-files/dockerhub.mk
include ./mk-files/semver.mk
include ./mk-files/release.mk
include ./mk-files/release-test.mk
include ./mk-files/release-notes.mk
include ./mk-files/unrelease.mk
include ./mk-files/utils.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
HOSTNAME := $(shell id -u -n)@$(shell hostname)
RESOLVED_PATH=github.com/confluentinc/cli/cmd/confluent

S3_BUCKET_PATH=s3://confluent.cloud
S3_STAG_FOLDER_NAME=cli-release-stag
S3_STAG_PATH=s3://confluent.cloud/$(S3_STAG_FOLDER_NAME)


.PHONY: clean
clean:
	rm -rf $(shell pwd)/dist

.PHONY: generate
generate:
	@go generate ./...

.PHONY: deps
deps:
	export GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc && \
	export GO111MODULE=on && \
	export GOPRIVATE=github.com/confluentinc && \
	go get github.com/goreleaser/goreleaser@v0.142.0 && \
	go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.21.0 && \
	go get github.com/mitchellh/golicense@v0.1.1

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

#
# START DEVELOPMENT HELPERS
# Usage: make run-ccloud -- version
#        make run-ccloud -- --version
#

# If the first argument is "run-ccloud"...
ifeq (run-ccloud,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run-ccloud"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

# If the first argument is "run-confluent"...
ifeq (run-confluent,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run-confluent"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: run-ccloud
run-ccloud:
	 @go run -ldflags '-buildmode=exe -X main.cliName=ccloud' cmd/confluent/main.go $(RUN_ARGS)

.PHONY: run-confluent
run-confluent:
	 @go run -ldflags '-buildmode=exe -X main.cliName=confluent' cmd/confluent/main.go $(RUN_ARGS)

#
# END DEVELOPMENT HELPERS
#

.PHONY: build
build:
	make build-ccloud
	make build-confluent

.PHONY: build-ccloud
build-ccloud:
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" goreleaser release --snapshot --rm-dist -f .goreleaser-ccloud$(GORELEASER_SUFFIX)

.PHONY: build-confluent
build-confluent:
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc GONOSUMDB=github.com/confluentinc,github.com/golangci/go-misc VERSION=$(VERSION) HOSTNAME="$(HOSTNAME)" goreleaser release --snapshot --rm-dist -f .goreleaser-confluent$(GORELEASER_SUFFIX)

.PHONY: build-integ
build-integ:
	make build-integ-nonrace
	make build-integ-race

.PHONY: build-integ-nonrace
build-integ-nonrace:
	make build-integ-ccloud-nonrace
	make build-integ-confluent-nonrace

.PHONY: build-integ-ccloud-nonrace
build-integ-ccloud-nonrace:
	binary="ccloud_test" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	GO111MODULE=on go test ./cmd/confluent -ldflags="-buildmode=exe -s -w -X $(RESOLVED_PATH).cliName=ccloud \
	-X $(RESOLVED_PATH).commit=$(REF) -X $(RESOLVED_PATH).host=$(HOSTNAME) -X $(RESOLVED_PATH).date=$(DATE) \
	-X $(RESOLVED_PATH).version=$(VERSION) -X $(RESOLVED_PATH).isTest=true" -tags testrunmain -coverpkg=./... -c -o $${binexe}

.PHONY: build-integ-confluent-nonrace
build-integ-confluent-nonrace:
	binary="confluent_test" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	GO111MODULE=on go test ./cmd/confluent -ldflags="-buildmode=exe -s -w -X $(RESOLVED_PATH).cliName=confluent \
		    -X $(RESOLVED_PATH).commit=$(REF) -X $(RESOLVED_PATH).host=$(HOSTNAME) -X $(RESOLVED_PATH).date=$(DATE) \
		    -X $(RESOLVED_PATH).version=$(VERSION) -X $(RESOLVED_PATH).isTest=true" -tags testrunmain -coverpkg=./... -c -o $${binexe}

.PHONY: build-integ-race
build-integ-race:
	make build-integ-ccloud-race
	make build-integ-confluent-race

.PHONY: build-integ-ccloud-race
build-integ-ccloud-race:
	binary="ccloud_test_race" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	GO111MODULE=on go test ./cmd/confluent -ldflags="-buildmode=exe -s -w -X $(RESOLVED_PATH).cliName=ccloud \
	-X $(RESOLVED_PATH).commit=$(REF) -X $(RESOLVED_PATH).host=$(HOSTNAME) -X $(RESOLVED_PATH).date=$(DATE) \
	-X $(RESOLVED_PATH).version=$(VERSION) -X $(RESOLVED_PATH).isTest=true" -tags testrunmain -coverpkg=./... -c -o $${binexe} -race

.PHONY: build-integ-confluent-race
build-integ-confluent-race:
	binary="confluent_test_race" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	GO111MODULE=on go test ./cmd/confluent -ldflags="-buildmode=exe -s -w -X $(RESOLVED_PATH).cliName=confluent \
		    -X $(RESOLVED_PATH).commit=$(REF) -X $(RESOLVED_PATH).host=$(HOSTNAME) -X $(RESOLVED_PATH).date=$(DATE) \
		    -X $(RESOLVED_PATH).version=$(VERSION) -X $(RESOLVED_PATH).isTest=true" -tags testrunmain -coverpkg=./... -c -o $${binexe} -race

# If you setup your laptop following https://github.com/confluentinc/cc-documentation/blob/master/Operations/Laptop%20Setup.md
# then assuming caas.sh lives here should be fine
define caasenv-authenticate
	source $$GOPATH/src/github.com/confluentinc/cc-dotfiles/caas.sh && caasenv prod
endef

.PHONY: fmt
fmt:
	@goimports -e -l -local github.com/confluentinc/cli/ -w $(ALL_SRC)

.PHONY: release-ci
release-ci:
ifneq ($(SEMAPHORE_GIT_PR_BRANCH),)
	true
else ifeq ($(SEMAPHORE_GIT_BRANCH),master)
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
	@GO111MODULE=on go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)

.PHONY: lint-go
lint-go:
	@GO111MODULE=on golangci-lint run --timeout=10m

.PHONY: lint
lint:
ifeq ($(shell uname),Darwin)
	true
else ifneq (,$(findstring NT,$(shell uname)))
	true
else
	make lint-go && \
	make lint-cli && \
	make lint-installers
endif

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
		[ -t 0 ] && args="" || args="-plain" ; \
		GITHUB_TOKEN=$(token) golicense $${args} .golicense.hcl ./dist/$${binary}/$(shell go env GOOS)_$(shell go env GOARCH)/$${binary} ; \
		echo ; \
	done

.PHONY: coverage-unit
coverage-unit:
      ifdef CI
	@# Run unit tests with coverage.
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc go test -v -race -coverpkg=$$(go list ./... | grep -v test | grep -v mock | tr '\n' ',' | sed 's/,$$//g') -coverprofile=unit_coverage.txt $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
	@grep -h -v "mode: atomic" unit_coverage.txt >> coverage.txt
      else
	@# Run unit tests.
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc go test -race -coverpkg=./... $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
      endif

.PHONY: coverage-integ
coverage-integ:
      ifdef CI
	@# Run integration tests with coverage.
	@GO111MODULE=on INTEG_COVER=on go test -v $$(go list ./... | grep cli/test) $(INT_TEST_ARGS) -timeout 20m -ldflags '-buildmode=exe'
	@grep -h -v "mode: atomic" integ_coverage.txt >> coverage.txt
      else
	@# Run integration tests.
	@GO111MODULE=on GOPRIVATE=github.com/confluentinc go test -v -race $$(go list ./... | grep cli/test) $(INT_TEST_ARGS) -timeout 20m -ldflags '-buildmode=exe'
      endif


.PHONY: test-prep
test-prep: lint
      ifdef CI
    @echo "mode: atomic" > coverage.txt
      endif

.PHONY: test
test: test-prep coverage-unit coverage-integ test-installers

.PHONY: unit-test
unit-test: test-prep coverage-unit

.PHONY: int-test
int-test: test-prep coverage-integ

.PHONY: doctoc
doctoc:
	npx doctoc README.md

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch
