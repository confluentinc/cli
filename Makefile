CCSTRUCTS = $(GOPATH)/src/github.com/confluentinc/cc-structs

PROTO = shared/connect

compile-proto:
	protoc -I $(PROTO) -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor $(PROTO)/*.proto --gogo_out=plugins=grpc:$(PROTO)

install-plugins:
	go install ./plugin/...

deps:
	which dep 2>/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure $(ARGS)

test:
	go test -v -cover $(TEST_ARGS) ./...

clean:
	rm $(PROTO)/*.pb.go
