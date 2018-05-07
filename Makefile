CCSTRUCTS = $(GOPATH)/src/github.com/confluentinc/cc-structs

PROTO = shared/connect

compile-proto:
	protoc -I $(PROTO) -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor $(PROTO)/*.proto --gogo_out=plugins=grpc:$(PROTO)

install-plugins:
	go install ./plugin/...

test:
	go test -v -cover $(TEST_ARGS) ./...

clean:
	rm $(PROTO)/*.pb.go
