ALL_SRC         := $(shell find . -name "*.go" | grep -v -e vendor)
GIT_REMOTE_NAME ?= origin

include ./semver.mk

.PHONY: deps
deps:
	@GO111MODULE=on go mod download
	@GO111MODULE=on go get github.com/golangci/golangci-lint/cmd/golangci-lint@v1.12.2

.PHONY: generate
generate:
	protoc shared/connect/*.proto -Ishared/connect -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/connect
	protoc shared/kafka/*.proto -Ishared/kafka -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/kafka
	protoc shared/ksql/*.proto -Ishared/ksql -I$(GOPATH)/src -I$(GOPATH)/src/github.com/confluentinc/ccloudapis --gogo_out=plugins=grpc:shared/ksql

.PHONY: install-plugins
install-plugins:
	@GO111MODULE=on go install ./dist/...

.PHONY: release
release: get-release-image commit-release tag-release

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
	@for d in $$(go list ./... | grep -v tools | grep -v vendor); do \
	  GO111MODULE=on go test -v -race -coverprofile=profile.out -covermode=atomic $$d || exit 2; \
	  if [ -f profile.out ]; then \
	    cat profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
      else
	@GO111MODULE=on go test -race -cover $(TEST_ARGS) $(go list ./... | grep -v tools | grep -v vendor)
      endif

.PHONY: test
test: lint coverage
