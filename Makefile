CCSTRUCTS := $(GOPATH)/src/github.com/confluentinc/cc-structs

.PHONY: deps
deps:
	which dep 2>/dev/null || go get -u github.com/golang/dep/cmd/dep
	which gometalinter 2>/dev/null || ( go get -u github.com/alecthomas/gometalinter && gometalinter --install &> /dev/null )
	dep ensure $(ARGS)

.PHONY: compile-proto
compile-proto:
	protoc -I shared/connect -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/connect/*.proto --gogo_out=plugins=grpc:shared/connect
	protoc -I shared/kafka -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/kafka/*.proto --gogo_out=plugins=grpc:shared/kafka

.PHONY: install-plugins
install-plugins:
	go install ./plugin/...

.PHONY: lint
lint:
	gometalinter ./... --vendor

.PHONY: test
test: lint
	go test -v -cover $(TEST_ARGS) ./...

.PHONY: clean
clean:
	rm $(PROTO)/*.pb.go
