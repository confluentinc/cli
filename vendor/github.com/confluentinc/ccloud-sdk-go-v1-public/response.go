package ccloud

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/gogo/protobuf/jsonpb"
	"github.com/gogo/protobuf/proto"
)

// JsonDecoder decodes http response JSON into a JSON-tagged struct value.
//
// This is copied from sling because they didn't export it and we need to "reset"
// the decoder to use basic json encoding sometimes (from a base sling instance)
type JsonDecoder struct {
}

// Decode decodes the Response Body into the value pointed to by v.
// Caller must provide a non-nil v and close the resp.Body.
func (d JsonDecoder) Decode(resp *http.Response, v interface{}) error {
	return json.NewDecoder(resp.Body).Decode(v)
}

// JSONPBDecoder deserializes response bodies as protobuf JSON messages by implementing sling.Decoder.
type JSONPBDecoder struct {
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
func (d JSONPBDecoder) Decode(resp *http.Response, v interface{}) error {
	if msg, ok := v.(proto.Message); ok {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		buf := bytes.NewBuffer(body)
		err = d.Unmarshal(buf, msg)
		if err != nil {
			err = fmt.Errorf("%s\n%s", err.Error(), string(body))
		}
		return err
	} else {
		fmt.Sprintln("non-protobuf interface v given. Decoding with json instead of jsonpb.")
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}

		err = json.Unmarshal(body, &v)
		if err != nil {
			err = fmt.Errorf("%s\n%s", err, string(body))
		}
		return err
	}
}
