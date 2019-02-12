ALL_SRC               := $(shell find . -name "*.go" | grep -v -e vendor)
GOLANGCI_LINT_VERSION := 1.12.2

include ./semver.mk

.PHONY: deps
deps:
	@which goreleaser >/dev/null 2>&1 || go get github.com/goreleaser/goreleaser >/dev/null 2>&1
	@(golangci-lint --version | grep $(GOLANGCI_LINT_VERSION)) >/dev/null 2>&1 || curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s -- -b $(GOPATH)/bin v$(GOLANGCI_LINT_VERSION) >/dev/null 2>&1
	@GO111MODULE=on go mod download >/dev/null 2>&1

.PHONY: generate
generate:
	protoc shared/connect/*.proto -Ishared/connect -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/connect
	protoc shared/kafka/*.proto -Ishared/kafka -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/kafka
	protoc shared/ksql/*.proto -Ishared/ksql -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/ksql

.PHONY: install-plugins
install-plugins:
	@GO111MODULE=on go install ./dist/...

ifeq ($(shell uname),Darwin)
GORELEASER_CONFIG ?= .goreleaser-mac.yml
else
GORELEASER_CONFIG ?= .goreleaser-linux.yml
endif

.PHONY: binary
binary:
	@GO111MODULE=on goreleaser release --snapshot --rm-dist -f $(GORELEASER_CONFIG)

.PHONY: release
release: get-release-image commit-release tag-release
	echo '$(RELEASE_SVG)' > release.svg
	git add release.svg
	goreleaser

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
	@GO111MODULE=on go test -race -cover $(TEST_ARGS) ./...
      endif

.PHONY: test
test: lint coverage
