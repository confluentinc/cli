SHELL              := /bin/bash
ALL_SRC            := $(shell find . -name "*.go" | grep -v -e vendor)
GORELEASER_VERSION := v1.15.2

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
	go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION) && \
	TAGS=$(TAGS) CGO_ENABLED=$(CGO_ENABLED) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) VERSION=$(VERSION) GOEXPERIMENT=boringcrypto goreleaser build -f .goreleaser-build.yml --clean --single-target --snapshot

include ./mk-files/cc-cli-service.mk
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

S3_BUCKET_PATH=s3://confluent.cloud
S3_STAG_FOLDER_NAME=cli-release-stag
S3_STAG_PATH=s3://confluent.cloud/$(S3_STAG_FOLDER_NAME)

.PHONY: clean
clean:
	@for dir in bin dist docs legal release-notes; do \
		[ -d $$dir ] && rm -r $$dir || true ; \
	done

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: lint
lint: lint-go lint-cli

.PHONY: lint-go
lint-go:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.51.1 && \
	golangci-lint run --enable dupword,gofmt,goimports,gomoddirectives,govet,ineffassign,misspell,prealloc,unparam,unused,usestdlibvars,whitespace --timeout=10m
	@echo "✅  golangci-lint"

.PHONY: lint-cli
lint-cli: cmd/lint/en_US.aff cmd/lint/en_US.dic
	go run cmd/lint/main.go -aff-file $(word 1,$^) -dic-file $(word 2,$^) $(ARGS)
	@echo "✅  cmd/lint/main.go"

cmd/lint/en_US.aff:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.aff?format=TEXT" | base64 -D > $@

cmd/lint/en_US.dic:
	curl -s "https://chromium.googlesource.com/chromium/deps/hunspell_dictionaries/+/master/en_US.dic?format=TEXT" | base64 -D > $@

.PHONY: unit-test
unit-test:
ifdef CI
	go install gotest.tools/gotestsum@v1.8.2 && \
	gotestsum --junitfile unit-test-report.xml -- -v -race -coverprofile coverage.out $$(go list ./... | grep -v test)
else
	go test -v $$(go list ./... | grep -v test) $(UNIT_TEST_ARGS)
endif

.PHONY: build-for-integration-test
build-for-integration-test:
ifdef CI
	go build -cover -ldflags="-s -w -X main.commit=$(REF) -X main.date=$(DATE) -X main.version=$(VERSION) -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
else
	go build -ldflags="-s -w -X main.commit=$(REF) -X main.date=$(DATE) -X main.version=$(VERSION) -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
endif


.PHONY: integration-test
integration-test:
ifdef CI
	go install gotest.tools/gotestsum@v1.8.2 && \
	export GOCOVERDIR=test/coverage && \
	if [ -d $${GOCOVERDIR} ]; then rm -r $${GOCOVERDIR}; fi && \
	mkdir $${GOCOVERDIR} && \
	gotestsum --junitfile integration-test-report.xml -- -v -race $$(go list ./... | grep test) && \
	go tool covdata textfmt -i $${GOCOVERDIR} -o test/coverage.out
else
	go test -v $$(go list ./... | grep test) $(INTEGRATION_TEST_ARGS)
endif

.PHONY: test
test: unit-test integration-test

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch
