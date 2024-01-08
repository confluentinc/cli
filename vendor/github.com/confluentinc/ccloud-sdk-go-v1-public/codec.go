package ccloud

import (
	"bytes"
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
	codec1978 "github.com/ugorji/go/codec"
)

// hack for dep so it doesn't remove the dependency
var _ = codec1978.GenVersion

func Marshal(v proto.Message) ([]byte, error) {
	return proto.Marshal(v)
}

func Unmarshal(b []byte, v proto.Message) error {
	return proto.Unmarshal(b, v)
}

var jsonMarshaler = jsonpb.Marshaler{
	OrigName:     true,
	EmitDefaults: true,
	EnumsAsInts:  false,
}

func MarshalJSONToBytes(v proto.Message) ([]byte, error) {
	s, err := jsonMarshaler.MarshalToString(v)
	if err != nil {
		return nil, err
	}
	return []byte(s), nil
}

func MarshalJSON(out io.Writer, v proto.Message) error {
	return jsonMarshaler.Marshal(out, v)
}

var jsonUnmarshaler = jsonpb.Unmarshaler{}

func UnmarshalJSON(in io.Reader, v proto.Message) error {
	return jsonUnmarshaler.Unmarshal(in, v)
}

func UnmarshalJSONBytes(b []byte, v proto.Message) error {
	buf := bytes.NewBufferString(string(b))
	return jsonUnmarshaler.Unmarshal(buf, v)
}

func WrapQuotes(str string) string {
	return fmt.Sprintf("\"%s\"", str)
}

func UnwrapQuotes(in string) (out string) {
	out = in
	if len(out) > 0 && out[0] == '"' {
		out = out[1:]
	}
	if len(out) > 0 && out[len(out)-1] == '"' {
		out = out[:len(out)-1]
	}
	return
}
