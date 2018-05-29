include ./mk-include/cc-semver.mk

CCSTRUCTS := $(GOPATH)/src/github.com/confluentinc/cc-structs

COMPONENTS := confluent-kafka-plugin confluent-connect-plugin
component = $(word 1, $@)

.PHONY: deps
deps:
	@which dep 2>/dev/null || go get -u github.com/golang/dep/cmd/dep
	@which gometalinter 2>/dev/null || ( go get -u github.com/alecthomas/gometalinter && gometalinter --install &> /dev/null )
	dep ensure $(ARGS)

.PHONY: compile-proto
compile-proto:
	protoc -I shared/connect -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/connect/*.proto --gogo_out=plugins=grpc:shared/connect
	protoc -I shared/kafka -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/kafka/*.proto --gogo_out=plugins=grpc:shared/kafka

.PHONY: install-plugins
install-plugins:
	go install ./plugin/...

.PHONY: cli
cli:
	@mkdir -p release/
	GOOS=linux GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/cli-$(VERSION)-linux-amd64 main.go
	GOOS=darwin GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/cli-$(VERSION)-darwin-amd64 main.go
	GOOS=windows GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/cli-$(VERSION)-windows-amd64 main.go

.PHONY: $(COMPONENTS)
$(COMPONENTS):
	@mkdir -p release/
	GOOS=linux GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/$(component)-$(VERSION)-linux-amd64 plugin/$(component)/main.go
	GOOS=darwin GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/$(component)-$(VERSION)-darwin-amd64 plugin/$(component)/main.go
	GOOS=windows GOARCH=amd64 go build -gcflags=-trimpath=$(GOPATH) -asmflags=-trimpath=$(GOPATH) -o release/$(component)-$(VERSION)-windows-amd64 plugin/$(component)/main.go

.PHONY: build
build: cli $(COMPONENTS)

.PHONY: release-s3
release-s3: build
	aws s3 cp release/cli-$(VERSION)-linux-amd64 s3://cloud-confluent-bin/cli/cli-$(VERSION)-linux-amd64
	aws s3 cp release/cli-$(VERSION)-darwin-amd64 s3://cloud-confluent-bin/cli/cli-$(VERSION)-darwin-amd64
	aws s3 cp release/cli-$(VERSION)-windows-amd64 s3://cloud-confluent-bin/cli/cli-$(VERSION)-windows-amd64
	for component in $(COMPONENTS) ; do \
		aws s3 cp release/$$component-$(VERSION)-linux-amd64 s3://cloud-confluent-bin/cli/components/$$component/$$component-$(VERSION)-linux-amd64 ; \
		aws s3 cp release/$$component-$(VERSION)-darwin-amd64 s3://cloud-confluent-bin/cli/components/$$component/$$component-$(VERSION)-darwin-amd64 ; \
		aws s3 cp release/$$component-$(VERSION)-windows-amd64 s3://cloud-confluent-bin/cli/components/$$component/$$component-$(VERSION)-windows-amd64 ; \
	done

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
