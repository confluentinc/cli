package ccstructs

import (
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

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

var jsonUnmarshaler = jsonpb.Unmarshaler{}

func UnmarshalJSON(in io.Reader, v proto.Message) error {
	return jsonUnmarshaler.Unmarshal(in, v)
}
