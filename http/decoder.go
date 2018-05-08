package http

import (
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// JSONPBDecoder implements sling.Decoder to decode JSON protobuf responses.
type JSONPBDecoder struct{
	jsonpb.Unmarshaler
}

// NewJSONPBDecoder returns a new JSONPBDecoder that's lenient for error responses.
func NewJSONPBDecoder() JSONPBDecoder {
	return JSONPBDecoder{
		jsonpb.Unmarshaler{
			// This is required to handle malformed tokens on regular (non-auth) requests
			AllowUnknownFields: true,
		},
	}
}

// Decode reads the next value from the reader and stores it in the value pointed to by v.
func (d JSONPBDecoder) Decode(r io.Reader, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		return d.Unmarshal(r, msg)
	}
	return fmt.Errorf("non-protobuf interface v given")
}
