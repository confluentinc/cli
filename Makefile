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
GORELEASER_SUFFIX ?= -mac.yml
else
GORELEASER_SUFFIX ?= -linux.yml
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

.PHONY: release
release: get-release-image commit-release tag-release
	make gorelease
	make publish

.PHONY: gorelease
gorelease:
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist .goreleaser-ccloud.yml
	@GO111MODULE=on VERSION=$(VERSION) HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist .goreleaser-confluent.yml

.PHONY: dist-ccloud
dist-ccloud:
	@# unfortunately goreleaser only supports one archive right now (either tar/zip or binaries): https://github.com/goreleaser/goreleaser/issues/705
	@# we had goreleaser upload binaries (they're uncompressed, so goreleaser's parallel uploads will save more time with binaries than archives)
	cp LICENSE dist/ccloud/darwin_amd64/
	cp LICENSE dist/ccloud/linux_amd64/
	cp LICENSE dist/ccloud/linux_386/
	cp LICENSE dist/ccloud/windows_amd64/
	cp LICENSE dist/ccloud/windows_amd64/
	cp INSTALL.md dist/ccloud/darwin_amd64/
	cp INSTALL.md dist/ccloud/linux_amd64/
	cp INSTALL.md dist/ccloud/linux_386/
	cp INSTALL.md dist/ccloud/windows_amd64/
	cp INSTALL.md dist/ccloud/windows_amd64/
	tar -czf dist/ccloud/ccloud_$(VERSION)_darwin_amd64.tar.gz -C dist/ccloud/darwin_amd64 .
	tar -czf dist/ccloud/ccloud_$(VERSION)_linux_amd64.tar.gz -C dist/ccloud/linux_amd64 .
	tar -czf dist/ccloud/ccloud_$(VERSION)_linux_386.tar.gz -C dist/ccloud/linux_386 .
	zip -jqr dist/ccloud/ccloud_$(VERSION)_windows_amd64.zip dist/ccloud/windows_amd64/*
	zip -jqr dist/ccloud/ccloud_$(VERSION)_windows_386.zip dist/ccloud/windows_386/*
	cp dist/ccloud/ccloud_$(VERSION)_darwin_amd64.tar.gz dist/ccloud/ccloud_latest_darwin_amd64.tar.gz
	cp dist/ccloud/ccloud_$(VERSION)_linux_amd64.tar.gz dist/ccloud/ccloud_latest_linux_amd64.tar.gz
	cp dist/ccloud/ccloud_$(VERSION)_linux_386.tar.gz dist/ccloud/ccloud_latest_linux_386.tar.gz
	cp dist/ccloud/ccloud_$(VERSION)_windows_amd64.zip dist/ccloud/ccloud_latest_windows_amd64.zip
	cp dist/ccloud/ccloud_$(VERSION)_windows_386.zip dist/ccloud/ccloud_latest_windows_386.zip

.PHONY: publish-ccloud
publish: dist-ccloud
	aws s3 cp dist/ccloud/ s3://confluent.cloud/ccloud-cli/archives/$(VERSION:v%=%)/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_latest_*" --acl public-read
	aws s3 cp dist/ccloud/ s3://confluent.cloud/ccloud-cli/archives/latest/ --recursive --exclude "*" --include "*.tar.gz" --include "*.zip" --exclude "*_$(VERSION)_*" --acl public-read

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
