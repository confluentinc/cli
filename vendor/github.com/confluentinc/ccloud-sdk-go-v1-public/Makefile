.PHONY: protoc-binaries
protoc-binaries:
	go install github.com/gogo/protobuf/protoc-gen-gogo@v1.3.2
	go install github.com/gogo/googleapis/protoc-gen-gogogoogleapis@v1.4.1
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.1
	go install github.com/ckaznocha/protoc-gen-lint@v0.2.4
	go install github.com/travisjeffery/proto-go-sql/protoc-gen-sql@v0.0.0-20190911121832-39ff47280e87
	go install github.com/confluentinc/proto-go-setter/protoc-gen-setter@v0.3.0
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-grpc-gateway@v1.9.5
	go install github.com/grpc-ecosystem/grpc-gateway/protoc-gen-swagger@v1.9.5
	# Install protoc-gen-structs, since we already have it just install it from local
	(cd protoc-gen-structs && go install .)
