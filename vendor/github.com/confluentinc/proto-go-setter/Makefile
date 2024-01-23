compile:
	protoc --gogo_out=Mgoogle/protobuf/descriptor.proto=github.com/gogo/protobuf/protoc-gen-gogo/descriptor:setter *.proto

build:
	go build -o bin/protoc-gen-setter ./protoc-gen-setter

examples: build
	$(eval PATH=$(PATH):bin/)
	protoc -I $(GOPATH)/src -I . --setter_out=. --gogo_out=. example/*.proto

test: examples
	go test ./...

.PHONY: compile build examples test
