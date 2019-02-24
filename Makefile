ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin

include ./semver.mk

REF := $(shell [ -d .git ] && git rev-parse --short HEAD || echo "none")
DATE := $(shell date -u)
HOSTNAME := $(shell id -u -n)@$(shell hostname -f)

.PHONY: clean
clean:
	rm -rf $(shell pwd)/dist

.PHONY: deps
deps:
	@GO111MODULE=on go mod download
	@GO111MODULE=on go mod verify
	@GO111MODULE=on go get github.com/goreleaser/goreleaser@v0.101.0
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.12.2
	@GO111MODULE=on go mod vendor

.PHONY: generate
generate:
	@GO111MODULE=on protoc shared/connect/*.proto -Ishared/connect -I$(shell pwd)/vendor/ -I$(shell pwd)/vendor/github.com/confluentinc/ccloudapis/ --gogo_out=plugins=grpc:shared/connect
	@GO111MODULE=on protoc shared/kafka/*.proto -Ishared/kafka -I$(shell pwd)/vendor/ -I$(shell pwd)/vendor/github.com/confluentinc/ccloudapis/ --gogo_out=plugins=grpc:shared/kafka
	@GO111MODULE=on protoc shared/ksql/*.proto -Ishared/ksql -I$(shell pwd)/vendor/ -I$(shell pwd)/vendor/github.com/confluentinc/ccloudapis/ --gogo_out=plugins=grpc:shared/ksql

build: generate build-go install-plugins

.PHONY: install-plugins
install-plugins:
	@GOBIN=$(shell pwd)/dist GO111MODULE=on go install ./plugin/...

ifeq ($(shell uname),Darwin)
GORELEASER_CONFIG ?= .goreleaser-mac.yml
else
GORELEASER_CONFIG ?= .goreleaser-linux.yml
endif

.PHONY: build-go
build-go:
	@GO111MODULE=on HOSTNAME=$(HOSTNAME) goreleaser release --rm-dist

.PHONY: release
release: get-release-image commit-release tag-release
	@GO111MODULE=on go mod vendor
# (cody+norwood): temporarily skipping validation to try to unbreak CI
	goreleaser release --skip-validate --skip-publish --rm-dist

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
