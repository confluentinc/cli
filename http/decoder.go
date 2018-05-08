package http

import (
	"fmt"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

type JSONPBDecoder struct{
	decoder jsonpb.Unmarshaler
}

func NewJSONPBDecoder() JSONPBDecoder {
	return JSONPBDecoder{
		decoder: jsonpb.Unmarshaler{
			// This is required to handle malformed tokens on regular (non-auth) requests
			AllowUnknownFields: true,
		},
	}
}

// Decode reads the next value from the reader and stores it in the value pointed to by v.
func (d JSONPBDecoder) Decode(r io.Reader, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		return d.decoder.Unmarshal(r, msg)
	}
	return fmt.Errorf("non-protobuf interface v given")
}
