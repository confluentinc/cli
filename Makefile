SHELL           := /bin/bash
ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin
MAIN_BRANCH     ?= main
RELEASE_BRANCH  ?= main

.PHONY: build # compile natively based on the system
build:
ifneq "" "$(findstring NT,$(shell uname))" # build for Windows
	CC=gcc CXX=g++ make cli-builder
else ifneq (,$(findstring Linux,$(shell uname)))
    ifneq (,$(findstring musl,$(shell ldd --version))) # build for musl Linux
		CC=gcc CXX=g++ TAGS=musl make cli-builder
    else # build for glibc Linux
		CC=gcc CXX=g++ make cli-builder
    endif
else
    ifneq (,$(findstring x86_64,$(shell uname -m))) # build for Darwin/amd64
		make cli-builder
    else # build for Darwin/arm64
		make switch-librdkafka-arm64
		make cli-builder || true 
		make restore-librdkafka-amd64
    endif
endif

.PHONY: cross-build # cross-compile from Darwin/amd64 machine to Win64, Linux64 and Darwin/arm64
cross-build:
ifeq ($(GOARCH),arm64) # build for darwin/arm64.
	make build-darwin-arm64
else # build for amd64 arch
    ifeq ($(GOOS),windows)
		CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_LDFLAGS="-static" make cli-builder
    else ifeq ($(GOOS),linux) 
		CGO_ENABLED=1 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl make cli-builder
    else # build for Darwin/amd64
		CGO_ENABLED=1 make cli-builder
    endif
endif

.PHONY: build-darwin-arm64
build-darwin-arm64:
	make switch-librdkafka-arm64
	CGO_ENABLED=1 make cli-builder || true 
	make restore-librdkafka-amd64

.PHONY: cli-builder
cli-builder:
	@GOPRIVATE=github.com/confluentinc TAGS=$(TAGS) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser build -f .goreleaser-build.yml --rm-dist --single-target --snapshot

include ./mk-files/dockerhub.mk
include ./mk-files/semver.mk
include ./mk-files/docs.mk
include ./mk-files/release.mk
include ./mk-files/release-test.mk
include ./mk-files/release-notes.mk
include ./mk-files/unrelease.mk
include ./mk-files/utils.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
HOSTNAME := $(shell id -u -n)@$(shell hostname)
RESOLVED_PATH=github.com/confluentinc/cli/cmd/confluent
RDKAFKA_VERSION = 1.8.2
RDKAFKA_PATH := $(shell find $(HOME)/git/go/1.17.6/pkg/mod/github.com/confluentinc -name confluent-kafka-go@v$(RDKAFKA_VERSION))/kafka/librdkafka_vendor

S3_BUCKET_PATH=s3://confluent.cloud
S3_STAG_FOLDER_NAME=cli-release-stag
S3_STAG_PATH=s3://confluent.cloud/$(S3_STAG_FOLDER_NAME)

.PHONY: clean
clean:
	@for dir in bin dist docs legal; do \
		[ -d $$dir ] && rm -r $$dir || true ; \
	done

.PHONY: generate
generate:
	@go generate ./...

.PHONY: deps
deps:
	go install github.com/goreleaser/goreleaser@v1.4.1 && \
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.44.0 && \
	go install github.com/mitchellh/golicense@v0.2.0

.PHONY: jenkins-deps
# Jenkins only depends on goreleaser, so we omit golangci-lint and golicense
jenkins-deps:
	go get github.com/goreleaser/goreleaser@v1.4.1

ifeq ($(shell uname),Darwin)
    SHASUM ?= gsha256sum
else ifneq (,$(findstring NT,$(shell uname)))
# TODO: I highly doubt this works. Completely untested. The output format is likely very different than expected.
    SHASUM ?= CertUtil SHA256 -hashfile
else ifneq (,$(findstring Windows,$(shell systeminfo)))
    SHASUM ?= CertUtil SHA256 -hashfile
else
    SHASUM ?= sha256sum
endif

show-args:
	@echo "VERSION: $(VERSION)"
	@echo "RDKAFKA_PATH: $(RDKAFKA_PATH)"

#
# START DEVELOPMENT HELPERS
# Usage: make run -- version
#        make run -- --version
#

# If the first argument is "run"...
ifeq (run,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  RUN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(RUN_ARGS):;@:)
endif

.PHONY: run
run:
	@GOPRIVATE=github.com/confluentinc go run cmd/confluent/main.go $(RUN_ARGS)

#
# END DEVELOPMENT HELPERS
#

.PHONY: build-integ
build-integ:
	make build-integ-nonrace
	make build-integ-race

.PHONY: build-integ-nonrace
build-integ-nonrace:
	binary="bin/confluent_test" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=$(REF) \
		-X $(RESOLVED_PATH).host=$(HOSTNAME) \
		-X $(RESOLVED_PATH).date=$(DATE) \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o $${binexe}

.PHONY: build-integ-race
build-integ-race:
	binary="bin/confluent_test_race" ; \
	[ "$${OS}" = "Windows_NT" ] && binexe=$${binary}.exe || binexe=$${binary} ; \
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=$(REF) \
		-X $(RESOLVED_PATH).host=$(HOSTNAME) \
		-X $(RESOLVED_PATH).date=$(DATE) \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o $${binexe} -race

# If you setup your laptop following https://github.com/confluentinc/cc-documentation/blob/master/Operations/Laptop%20Setup.md
# then assuming caas.sh lives here should be fine
define aws-authenticate
	source $$GOPATH/src/github.com/confluentinc/cc-dotfiles/caas.sh && eval $$(gimme-aws-creds --output-format export --roles "arn:aws:iam::050879227952:role/administrator")
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

.PHONY: lint
lint:
ifeq ($(shell uname),Darwin)
	true
else ifneq (,$(findstring NT,$(shell uname)))
	true
else
	@echo "Linting..."
	@make lint-go
	@make lint-cli
	@make lint-installers
endif

.PHONY: lint-go
lint-go:
	@golangci-lint run --timeout=10m
	@echo "✅  golangci-lint"

.PHONY: lint-cli
lint-cli: cmd/lint/en_US.aff cmd/lint/en_US.dic
	@go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)
	@echo "✅  cmd/lint/main.go"

cmd/lint/en_US.aff:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.aff?format=TEXT" | base64 -D > $@

cmd/lint/en_US.dic:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.dic?format=TEXT" | base64 -D > $@

.PHONY: lint-installers
lint-installers:
	@diff install-c* | grep -v -E "^---|^[0-9c0-9]|PROJECT_NAME|BINARY" && echo "diff between install scripts" && exit 1 || exit 0
	@echo "✅  installation script linter"

.PHONY: lint-licenses
## Scan and validate third-party dependency licenses
lint-licenses: build
	$(eval token := $(shell (grep github.com ~/.netrc -A 2 | grep password || grep github.com ~/.netrc -A 2 | grep login) | head -1 | awk -F' ' '{ print $$2 }'))
	echo Licenses for confluent binary ; \
	[ -t 0 ] && args="" || args="-plain" ; \
	GITHUB_TOKEN=$(token) golicense $${args} .golicense.hcl ./dist/confluent_$(shell go env GOOS)_$(shell go env GOARCH)/confluent || true

.PHONY: coverage-unit
## disabled -race flag for Windows build because of 'ThreadSanitizer failed to allocate' error: https://github.com/golang/go/issues/46099. Will renable in the future when this issue is resolved.
coverage-unit:
ifdef CI
	@# Run unit tests with coverage.
  ifeq "$(OS)" "Windows_NT"
	@GOPRIVATE=github.com/confluentinc go test -v -coverpkg=$$(go list ./... | grep -v test | grep -v mock | tr '\n' ',' | sed 's/,$$//g') -coverprofile=unit_coverage.txt $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
  else
	@GOPRIVATE=github.com/confluentinc go test -v -race -coverpkg=$$(go list ./... | grep -v test | grep -v mock | tr '\n' ',' | sed 's/,$$//g') -coverprofile=unit_coverage.txt $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
  endif
	@grep -h -v "mode: atomic" unit_coverage.txt >> coverage.txt
else
	@# Run unit tests.
  ifeq "$(OS)" "Windows_NT"
	@GOPRIVATE=github.com/confluentinc go test -coverpkg=./... $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
  else
	@GOPRIVATE=github.com/confluentinc go test -race -coverpkg=./... $$(go list ./... | grep -v vendor | grep -v test) $(UNIT_TEST_ARGS) -ldflags '-buildmode=exe'
  endif
endif

.PHONY: coverage-integ
coverage-integ:
      ifdef CI
	@# Run integration tests with coverage.
	@INTEG_COVER=on go test -v $$(go list ./... | grep cli/test) $(INT_TEST_ARGS) -timeout 45m
	@grep -h -v "mode: atomic" integ_coverage.txt >> coverage.txt
      else
	@# Run integration tests.
	@GOPRIVATE=github.com/confluentinc go test -v -race $$(go list ./... | grep cli/test) $(INT_TEST_ARGS) -timeout 45m
      endif

.PHONY: test-prep
test-prep: lint
      ifdef CI
    @echo "mode: atomic" > coverage.txt
      endif

.PHONY: test
test: test-prep coverage-unit coverage-integ test-installer

.PHONY: unit-test
unit-test: test-prep coverage-unit

.PHONY: int-test
int-test: test-prep coverage-integ

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch
