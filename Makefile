include ./mk-include/cc-semver.mk

CCSTRUCTS := $(GOPATH)/src/github.com/confluentinc/cc-structs

COMPONENTS := confluent-kafka-plugin confluent-connect-plugin
component = $(word 1, $@)

GO_LDFLAGS := -X main.version=$(VERSION)
GO_GCFLAGS := -trimpath=$(GOPATH)
GO_ASMFLAGS := -trimpath=$(GOPATH)

RELEASE_ARCH := 386 amd64
RELEASE_OS := linux darwin windows
RELEASE_OSARCH := !darwin/386

.PHONY: deps
deps:
	@which dep >/dev/null 2>&1 || go get github.com/golang/dep/cmd/dep
	@which gometalinter >/dev/null 2>&1 || ( go get github.com/alecthomas/gometalinter && gometalinter --install &> /dev/null )
	@which gox >/dev/null 2>&1 || go get github.com/mitchellh/gox
	dep ensure $(ARGS)

.PHONY: compile-proto
compile-proto:
	protoc -I shared/connect -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/connect/*.proto --gogo_out=plugins=grpc:shared/connect
	protoc -I shared/kafka -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/kafka/*.proto --gogo_out=plugins=grpc:shared/kafka

.PHONY: install-plugins
install-plugins:
	go install ./plugin/...

.PHONY: dev
dev:
	@gox -os="$(shell go env GOOS)" -arch="$(shell go env GOARCH)" \
	  -ldflags="$(GO_LDFLAGS)" -gcflags="$(GO_GCFLAGS)" -asmflags="$(GO_ASMFLAGS)" \
	  -output="{{if eq .Dir \"cli\"}}confluent{{else}}{{.Dir}}{{end}}" ./...

.PHONY: build
build:
	@gox -os="$(RELEASE_OS)" -arch="$(RELEASE_ARCH)" -osarch="$(RELEASE_OSARCH)" \
	  -ldflags="$(GO_LDFLAGS)" -gcflags="$(GO_GCFLAGS)" -asmflags="$(GO_ASMFLAGS)" \
	  -output="build/$(VERSION)/{{.OS}}_{{.Arch}}/{{if eq .Dir \"cli\"}}confluent{{else}}{{.Dir}}{{end}}" ./...

.PHONY: release-s3
release-s3:
	aws s3 sync build/$(VERSION)/ s3://cloud-confluent-bin/cli/$(VERSION)/

.PHONY: release
release: get-release-image commit-release tag-release
	make release-s3

.PHONY: release-ci
release-ci:
ifeq ($(BRANCH_NAME),master)
	make release
else
	true
endif

.PHONY: lint
lint:
	gometalinter ./... --vendor

.PHONY: coverage
coverage:
      ifdef CI
	@echo "" > coverage.txt
	@for d in $$(go list ./... | grep -v vendor); do \
	  go test -v -race -coverprofile=profile.out -covermode=atomic $$d || exit 2; \
	  if [ -f profile.out ]; then \
	    cat profile.out >> coverage.txt; \
	    rm profile.out; \
	  fi; \
	done
      else
	@go test -race -cover $(TEST_ARGS) ./...
      endif

.PHONY: test
test: lint coverage

.PHONY: clean
clean:
	rm $(PROTO)/*.pb.go
