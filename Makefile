### BEGIN HEADERS ###
# This block is managed by ServiceBot plugin - Make. The content in this block is created using a common
# template and configurations in service.yml.
# Modifications in this block will be overwritten by generated content in the nightly run.
# For more information, please refer to the page:
# https://confluentinc.atlassian.net/wiki/spaces/Foundations/pages/2871328913/Add+Make
SERVICE_NAME := cli
SERVICE_DEPLOY_NAME := cli

### END HEADERS ###
### BEGIN MK-INCLUDE UPDATE ###
### This section is managed by service-bot and should not be edited here.
### You can make changes upstream in https://github.com/confluentinc/cc-service-bot

CURL ?= curl
FIND ?= find
TAR ?= tar

# Mount netrc so curl can work from inside a container
DOCKER_NETRC_MOUNT ?= 1

GITHUB_API = api.github.com
GITHUB_MK_INCLUDE_OWNER := confluentinc
GITHUB_MK_INCLUDE_REPO := cc-mk-include
GITHUB_API_CC_MK_INCLUDE := https://$(GITHUB_API)/repos/$(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO)
GITHUB_API_CC_MK_INCLUDE_TARBALL := $(GITHUB_API_CC_MK_INCLUDE)/tarball
GITHUB_API_CC_MK_INCLUDE_VERSION ?= $(GITHUB_API_CC_MK_INCLUDE_TARBALL)/$(MK_INCLUDE_VERSION)

MK_INCLUDE_DIR := mk-include
MK_INCLUDE_LOCKFILE := .mk-include-lockfile
MK_INCLUDE_TIMESTAMP_FILE := .mk-include-timestamp
# For optimum performance, you should override MK_INCLUDE_TIMEOUT_MINS above the managed section headers to be
# a little longer than the worst case cold build time for this repo.
MK_INCLUDE_TIMEOUT_MINS ?= 240
# If this latest validated release is breaking you, please file a ticket with DevProd describing the issue, and
# if necessary you can temporarily override MK_INCLUDE_VERSION above the managed section headers until the bad
# release is yanked.
MK_INCLUDE_VERSION ?= v0.1452.0

# Make sure we always have a copy of the latest cc-mk-include release less than $(MK_INCLUDE_TIMEOUT_MINS) old:
# Note: The simply-expanded make variable makes sure this is run once per make invocation.
UPDATE_MK_INCLUDE := $(shell \
	func_fatal() { echo "$$*" >&2; echo output here triggers error below; exit 1; } ; \
	test -z "`git ls-files $(MK_INCLUDE_DIR)`" || { \
		func_fatal 'fatal: checked in $(MK_INCLUDE_DIR)/ directory is preventing make from fetching recent cc-mk-include releases for CI'; \
	} ; \
	trap "rm -f $(MK_INCLUDE_LOCKFILE); exit" 0 2 3 15; \
	waitlock=0; while ! ( set -o noclobber; echo > $(MK_INCLUDE_LOCKFILE) ); do \
	   sleep $$waitlock; waitlock=`expr $$waitlock + 1`; \
	   test 14 -lt $$waitlock && { \
	      echo 'stealing stale lock after 105s' >&2; \
	      break; \
	   } \
	done; \
	test -s $(MK_INCLUDE_TIMESTAMP_FILE) || rm -f $(MK_INCLUDE_TIMESTAMP_FILE); \
	{ test -d $(MK_INCLUDE_DIR) && test -d /proc && test -z "$(cat /proc/1/sched 2>&1 |head -n 1 |grep init)"; } || \
	test -z "`$(FIND) $(MK_INCLUDE_TIMESTAMP_FILE) -mmin +$(MK_INCLUDE_TIMEOUT_MINS) 2>&1`" || { \
	   GHAUTH=$$(grep -sq 'machine $(GITHUB_API)' ~/.netrc && echo netrc || \
	     ( command -v gh > /dev/null && gh auth status -h github.com > /dev/null && echo gh )); \
	   test -n "$$GHAUTH" || \
	     func_fatal 'error: no GitHub token available via "~/.netrc" or "gh auth status".\nFollow https://confluentinc.atlassian.net/l/cp/0WXXRLDh to setup GitHub authentication.\n'; \
	   echo "downloading $(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO) $(MK_INCLUDE_VERSION) using $$GHAUTH" >&2; \
	   if [ "netrc" = "$$GHAUTH" ]; then \
	     $(CURL) --fail --silent --netrc --location "$(GITHUB_API_CC_MK_INCLUDE_VERSION)" --output $(MK_INCLUDE_TIMESTAMP_FILE)T --write-out '$(GITHUB_API_CC_MK_INCLUDE_VERSION): %{errormsg}\n' >&2; \
	   else \
	     gh release download --clobber --repo=$(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO) \
	        --archive=tar.gz --output $(MK_INCLUDE_TIMESTAMP_FILE)T $(MK_INCLUDE_VERSION) >&2; \
	   fi \
	   && TMP_MK_INCLUDE_DIR=$$(mktemp -d -t cc-mk-include.XXXXXXXXXX) \
	   && $(TAR) -C $$TMP_MK_INCLUDE_DIR --strip-components=1 -zxf $(MK_INCLUDE_TIMESTAMP_FILE)T \
	   && rm -rf $$TMP_MK_INCLUDE_DIR/tests \
	   && rm -rf $(MK_INCLUDE_DIR) \
	   && mv $$TMP_MK_INCLUDE_DIR $(MK_INCLUDE_DIR) \
	   && mv -f $(MK_INCLUDE_TIMESTAMP_FILE)T $(MK_INCLUDE_TIMESTAMP_FILE) \
	   && echo 'installed cc-mk-include $(MK_INCLUDE_VERSION) from $(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO)' >&2 \
	   || func_fatal unable to install cc-mk-include $(MK_INCLUDE_VERSION) from $(GITHUB_MK_INCLUDE_OWNER)/$(GITHUB_MK_INCLUDE_REPO) \
	   ; \
	} || { \
	   rm -f $(MK_INCLUDE_TIMESTAMP_FILE)T; \
	   if test -f $(MK_INCLUDE_TIMESTAMP_FILE); then \
	      touch $(MK_INCLUDE_TIMESTAMP_FILE); \
	      func_fatal 'unable to access $(GITHUB_MK_INCLUDE_REPO) fetch API to check for latest release; next try in $(MK_INCLUDE_TIMEOUT_MINS) minutes'; \
	   else \
	      func_fatal 'unable to access $(GITHUB_MK_INCLUDE_REPO) fetch API to bootstrap mk-include subdirectory'; \
	   fi; \
	} \
)
ifneq ($(UPDATE_MK_INCLUDE),)
    $(error mk-include update failed)
endif

# Export the (empty) .mk-include-check-FORCE target to allow users to trigger the mk-include
# download code above via make but without having to run any of the other targets, e.g. build.
.PHONY: .mk-include-check-FORCE
.mk-include-check-FORCE:
	@echo -n ""
### END MK-INCLUDE UPDATE ###
### BEGIN INCLUDES ###
# This block is managed by ServiceBot plugin - Make. The content in this block is created using a common
# template and configurations in service.yml.
# Modifications in this block will be overwritten by generated content in the nightly run. To include
# additional mk files, please add them before or after this generated block.
# For more information, please refer to the page:
# https://confluentinc.atlassian.net/wiki/spaces/Foundations/pages/2871328913/Add+Make
include ./mk-include/cc-begin.mk
include ./mk-include/cc-semver.mk
include ./mk-include/cc-semaphore.mk
include ./mk-include/cc-go.mk
include ./mk-include/cc-testbreak.mk
include ./mk-include/cc-vault.mk
include ./mk-include/cc-sonarqube.mk
include ./mk-include/cc-end.mk
### END INCLUDES ###
SHELL := /bin/bash
GORELEASER_VERSION := v1.21.2

# Compile natively based on the current system
.PHONY: build
build:
ifneq "" "$(findstring NT,$(shell uname))" # windows
	CC=gcc CXX=g++ $(MAKE) cli-builder
else ifneq (,$(findstring Linux,$(shell uname)))
	ifneq (,$(findstring musl,$(shell ldd --version))) # linux (musl)
		CC=gcc CXX=g++ TAGS=musl $(MAKE) cli-builder
	else # linux (glibc)
		CC=gcc CXX=g++ $(MAKE) cli-builder
	endif
else # darwin
	$(MAKE) cli-builder
endif

# Cross-compile from darwin to any of the OS/Arch pairs below
.PHONY: cross-build
cross-build:
ifeq ($(GOARCH),arm64)
	ifeq ($(GOOS),linux) # linux/arm64
		CC=aarch64-linux-musl-gcc CXX=aarch64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl $(MAKE) cli-builder
	else # darwin/arm64
		$(MAKE) cli-builder
	endif
else
	ifeq ($(GOOS),windows) # windows/amd64
		CC=x86_64-w64-mingw32-gcc CXX=x86_64-w64-mingw32-g++ CGO_LDFLAGS="-fstack-protector -static" $(MAKE) cli-builder
	else ifeq ($(GOOS),linux) # linux/amd64
		CC=x86_64-linux-musl-gcc CXX=x86_64-linux-musl-g++ CGO_LDFLAGS="-static" TAGS=musl $(MAKE) cli-builder
	else # darwin/amd64
		$(MAKE) cli-builder
	endif
endif

.PHONY: cli-builder
cli-builder:
	GOOS="" GOARCH="" CC="" CXX="" CGO_LDFLAGS="" go install github.com/goreleaser/goreleaser@$(GORELEASER_VERSION)

ifeq ($(GOLANG_FIPS),1)
	wget "https://go.dev/dl/go$$(cat .go-version).src.tar.gz" && \
	tar -xf go$$(cat .go-version).src.tar.gz && \
	git clone --branch go$$(cat .go-version)-1-openssl-fips --depth 1 https://github.com/golang-fips/go.git go-openssl && \
	cd go/ && \
	cat ../go-openssl/patches/*.patch | patch -p1 && \
	sed -i '' 's/linux/darwin/' src/crypto/internal/backend/nobackend.go && \
	sed -i '' 's/linux/darwin/' src/crypto/internal/backend/openssl.go && \
	sed -i '' 's/"libcrypto.so.%s"/"libcrypto.%s.dylib"/' src/crypto/internal/backend/openssl.go && \
	cd src/ && \
	./make.bash && \
	cd ../../
	PATH=$$(pwd)/go/bin:$$PATH GOROOT=$$(pwd)/go TAGS=$(TAGS) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) goreleaser build --clean --single-target --snapshot
	rm -rf go go-openssl go$$(cat .go-version).src.tar.gz
else
	TAGS=$(TAGS) CC=$(CC) CXX=$(CXX) CGO_LDFLAGS=$(CGO_LDFLAGS) goreleaser build --clean --single-target --snapshot
endif



.PHONY: clean
clean:
	for dir in bin dist docs legal prebuilt release-notes; do \
		[ -d $$dir ] && rm -r $$dir || true; \
	done

.PHONY: lint
lint: lint-go lint-cli

.PHONY: lint-go
lint-go:
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.8 && \
	golangci-lint run --timeout 10m
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
	go install gotest.tools/gotestsum@v1.12.1 && \
	gotestsum --junitfile unit-test-report.xml -- -timeout 0 -v -race -coverprofile=coverage.unit.out -covermode=atomic $$(go list ./... | grep -v github.com/confluentinc/cli/v4/test)
else
	go test -timeout 0 -v -coverprofile=coverage.unit.out -covermode=atomic $$(go list ./... | grep -v github.com/confluentinc/cli/v4/test) $(UNIT_TEST_ARGS)
endif

.PHONY: build-for-integration-test
build-for-integration-test:
ifdef CI
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
else
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent ./cmd/confluent
endif

.PHONY: build-for-integration-test-windows
build-for-integration-test-windows:
ifdef CI
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent.exe ./cmd/confluent
else
	go build -cover -ldflags="-s -w -X main.commit="00000000" -X main.date="1970-01-01T00:00:00Z" -X main.isTest=true" -o test/bin/confluent.exe ./cmd/confluent
endif

.PHONY: integration-test
integration-test:
ifdef CI
	go install gotest.tools/gotestsum@v1.12.1 && \
	export GOCOVERDIR=test/coverage && \
	rm -rf $${GOCOVERDIR} && mkdir $${GOCOVERDIR} && \
	gotestsum --junitfile integration-test-report.xml -- -timeout 0 -v -race $$(go list ./... | grep github.com/confluentinc/cli/v4/test) && \
	go tool covdata textfmt -i $${GOCOVERDIR} -o coverage.integration.out
else
	export GOCOVERDIR=test/coverage && \
	rm -rf $${GOCOVERDIR} && mkdir $${GOCOVERDIR} && \
	go test -timeout 0 -v $$(go list ./... | grep github.com/confluentinc/cli/v4/test) $(INTEGRATION_TEST_ARGS) && \
	go tool covdata textfmt -i $${GOCOVERDIR} -o coverage.integration.out
endif

.PHONY: test
test: unit-test integration-test

.PHONY: generate-packaging-patch
generate-packaging-patch:
	diff -u Makefile debian/Makefile | sed "1 s_Makefile_cli/Makefile_" > debian/patches/standard_build_layout.patch

.PHONY: coverage
coverage: ## Merge coverage data from unit and integration tests into coverage.txt
	@echo "Merging coverage data..."
	@echo "mode: atomic" > coverage.txt
	@tail -n +2 coverage.unit.out >> coverage.txt
	@tail -n +2 coverage.integration.out >> coverage.txt
	@echo "Coverage data saved to: coverage.txt"
	@artifact push workflow coverage.txt
