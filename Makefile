CCSTRUCTS = $(GOPATH)/src/github.com/confluentinc/cc-structs

compile-proto:
	protoc -I shared/connect -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/connect/*.proto --gogo_out=plugins=grpc:shared/connect
	protoc -I shared/kafka -I $(CCSTRUCTS) -I $(CCSTRUCTS)/vendor shared/kafka/*.proto --gogo_out=plugins=grpc:shared/kafka

install-plugins:
	go install ./plugin/...

deps:
	which dep 2>/dev/null || go get -u github.com/golang/dep/cmd/dep
	dep ensure $(ARGS)

test:
	go test -v -cover $(TEST_ARGS) ./...

clean:
	rm $(PROTO)/*.pb.go
