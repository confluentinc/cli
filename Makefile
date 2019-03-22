ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin
MASTER_BRANCH   ?= master
RELEASE_BRANCH  ?= master

include ./semver.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u)
HOSTNAME := $(shell id -u -n)@$(shell hostname -f)

.PHONY: clean
clean:
	rm -rf $(shell pwd)/dist

.PHONY: deps
deps:
	@GO111MODULE=on go get github.com/goreleaser/goreleaser@v0.101.0
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.12.2

build: build-go

ifeq ($(shell uname),Darwin)
GORELEASER_CONFIG ?= .goreleaser-mac.yml
else
GORELEASER_CONFIG ?= .goreleaser-linux.yml
endif

show-args:
	@echo "VERSION: $(VERSION)"

.PHONY: build-go
build-go:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --snapshot --rm-dist -f $(GORELEASER_CONFIG)

.PHONY: release
release: get-release-image commit-release tag-release
	make gorelease
	make publish

.PHONY: gorelease
gorelease:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist

.PHONY: dist
dist:
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	cp LICENSE dist/darwin_amd64/
	cp LICENSE dist/linux_amd64/
	cp LICENSE dist/linux_386/
	cp LICENSE dist/windows_amd64/
	cp LICENSE dist/windows_amd64/
	cp INSTALL.md dist/darwin_amd64/
	cp INSTALL.md dist/linux_amd64/
	cp INSTALL.md dist/linux_386/
	cp INSTALL.md dist/windows_amd64/
	cp INSTALL.md dist/windows_amd64/
	tar -czf dist/ccloud_$(VERSION)_darwin_amd64.tar.gz -C dist/darwin_amd64 .
	tar -czf dist/ccloud_$(VERSION)_linux_amd64.tar.gz -C dist/linux_amd64 .
	tar -czf dist/ccloud_$(VERSION)_linux_386.tar.gz -C dist/linux_386 .
	zip -jqr dist/ccloud_$(VERSION)_windows_amd64.zip dist/windows_amd64/*
	zip -jqr dist/ccloud_$(VERSION)_windows_386.zip dist/windows_386/*
	cp dist/ccloud_$(VERSION)_darwin_amd64.tar.gz dist/ccloud_latest_darwin_amd64.tar.gz
	cp dist/ccloud_$(VERSION)_linux_amd64.tar.gz dist/ccloud_latest_linux_amd64.tar.gz
	cp dist/ccloud_$(VERSION)_linux_386.tar.gz dist/ccloud_latest_linux_386.tar.gz
	cp dist/ccloud_$(VERSION)_windows_amd64.zip dist/ccloud_latest_windows_amd64.zip
	cp dist/ccloud_$(VERSION)_windows_386.zip dist/ccloud_latest_windows_386.zip

.PHONY: publish
publish: dist
	aws s3 cp dist/ s3://confluent.cloud/ccloud-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_latest_*" --acl public-read
	aws s3 cp dist/ s3://confluent.cloud/ccloud-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_$(VERSION)_*" --acl public-read

.PHONY: fmt
fmt:
	@gofmt -e -s -l -w $(ALL_SRC)

.PHONY: release-ci
release-ci:
ifeq ($(BRANCH_NAME),master)
	make release
else
	true
endif

.PHONY: lint
lint:
	@GO111MODULE=on golangci-lint run

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

.PHONY: test
test: lint coverage
