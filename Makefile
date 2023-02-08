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
else # build for Darwin
	make cli-builder
endif

.PHONY: cross-build # cross-compile from Darwin/amd64 machine to Win64, Linux64 and Darwin/arm64
cross-build:
ifeq ($(GOARCH),arm64)
    ifeq ($(GOOS),linux)
		CGO_ENABLED=1 CC=aarch64-linux-musl-gcc CXX=aarch64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl make cli-builder
    else # build for darwin/arm64
		CGO_ENABLED=1 make cli-builder
    endif
else # build for amd64 arch
    ifeq ($(GOOS),windows)
		CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_LDFLAGS="-static" make cli-builder
    else ifeq ($(GOOS),linux) 
		CGO_ENABLED=1 CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl make cli-builder
    else # build for Darwin/amd64
		CGO_ENABLED=1 make cli-builder
    endif
endif

.PHONY: cli-builder
cli-builder:
	@TAGS=$(TAGS) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) VERSION=$(VERSION) GOEXPERIMENT=boringcrypto goreleaser build -f .goreleaser-build.yml --rm-dist --single-target --snapshot

include ./mk-files/dockerhub.mk
include ./mk-files/semver.mk
include ./mk-files/docs.mk
include ./mk-files/release.mk
include ./mk-files/release-test.mk
include ./mk-files/release-notes.mk
include ./mk-files/unrelease.mk
include ./mk-files/usage.mk
include ./mk-files/utils.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
RESOLVED_PATH=github.com/confluentinc/cli/cmd/confluent

S3_BUCKET_PATH=s3://confluent.cloud
S3_STAG_FOLDER_NAME=cli-release-stag
S3_STAG_PATH=s3://confluent.cloud/$(S3_STAG_FOLDER_NAME)

.PHONY: clean
clean:
	@for dir in bin dist docs legal release-notes; do \
		[ -d $$dir ] && rm -r $$dir || true ; \
	done

.PHONY: deps
deps:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1 && \
	go install github.com/google/go-licenses@v1.5.0 && \
	go install github.com/goreleaser/goreleaser@v1.14.1 && \
	go install gotest.tools/gotestsum@v1.8.2

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: build-integ-nonrace
build-integ-nonrace:
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=$(REF) \
		-X $(RESOLVED_PATH).date=$(DATE) \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o bin/confluent_test

.PHONY: build-integ-race
build-integ-race:
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=$(REF) \
		-X $(RESOLVED_PATH).date=$(DATE) \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o bin/confluent_test_race -race

.PHONY: build-integ-nonrace-windows
build-integ-nonrace-windows:
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=12345678 \
		-X $(RESOLVED_PATH).date=2000-01-01T00:00:00Z \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o bin/confluent_test.exe

.PHONY: build-integ-race-windows
build-integ-race-windows:
	go test ./cmd/confluent -ldflags="-s -w \
		-X $(RESOLVED_PATH).commit=12345678 \
		-X $(RESOLVED_PATH).date=2000-01-01T00:00:00Z \
		-X $(RESOLVED_PATH).version=$(VERSION) \
		-X $(RESOLVED_PATH).isTest=true" \
		-tags testrunmain -coverpkg=./... -c -o bin/confluent_test_race.exe -race

# If you setup your laptop following https://github.com/confluentinc/cc-documentation/blob/master/Operations/Laptop%20Setup.md
# then assuming caas.sh lives here should be fine
define aws-authenticate
	source ~/git/go/src/github.com/confluentinc/cc-dotfiles/caas.sh && if ! aws sts get-caller-identity; then eval $$(gimme-aws-creds --output-format export --roles "arn:aws:iam::050879227952:role/administrator"); fi
endef

.PHONY: fmt
fmt:
	@goimports -e -l -local github.com/confluentinc/cli/ -w $(ALL_SRC)

.PHONY: lint
lint:
	make lint-go
	make lint-cli

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

.PHONY: lint-licenses
lint-licenses:
	go-licenses check ./...

.PHONY: unit-test
unit-test:
ifdef CI
	gotestsum --junitfile unit-test-report.xml -- -v -race $$(go list ./... | grep -v test)
else
	go test -v -race $$(go list ./... | grep -v test) $(UNIT_TEST_ARGS)
endif

.PHONY: int-test
int-test:
ifdef CI
	gotestsum --junitfile integration-test-report.xml -- -v -race $$(go list ./... | grep test)
else
	go test -v -race $$(go list ./... | grep test) $(INT_TEST_ARGS)
endif

.PHONY: test
test: unit-test int-test

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch
