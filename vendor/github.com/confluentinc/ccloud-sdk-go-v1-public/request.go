package ccloud

import (
	"bytes"
	"io"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// JSONPBBodyProvider serializes request bodies as protobuf JSON messages by implementing sling.BodyProvider.
//
// This encodes a JSON tagged struct value as a Body for requests.
// See https://godoc.org/github.com/golang/protobuf/jsonpb#Marshaler for details.
type JSONPBBodyProvider struct {
	body       proto.Message
	marshaller jsonpb.Marshaler
}

// JSONPBBodyProvider returns a new JSONPBBodyProvider
func NewJSONPBBodyProvider(body proto.Message) *JSONPBBodyProvider {
	return &JSONPBBodyProvider{
		body:       body,
		marshaller: jsonpb.Marshaler{},
	}
}

func (p *JSONPBBodyProvider) ContentType() string {
	return "application/json"
}

func (p *JSONPBBodyProvider) Body() (io.Reader, error) {
	buf := &bytes.Buffer{}
	err := p.marshaller.Marshal(buf, p.body)
	if err != nil {
		return nil, err
	}
	return buf, nil
}
