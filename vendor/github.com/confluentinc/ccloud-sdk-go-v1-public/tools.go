//go:build tools

package tools

import (
	_ "github.com/confluentinc/proto-go-setter"
	_ "github.com/envoyproxy/protoc-gen-validate"
	_ "github.com/gogo/googleapis/google/rpc"
	_ "github.com/gogo/googleapis/protoc-gen-gogogoogleapis"
	_ "github.com/gogo/protobuf/gogoproto"
	_ "github.com/gogo/protobuf/proto"
	_ "github.com/travisjeffery/mocker/cmd/mocker"
	_ "github.com/travisjeffery/proto-go-sql"
	_ "k8s.io/api"
)
