// +build launchdarkly_easyjson

package lduser

import (
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"github.com/mailru/easyjson/jlexer"
	ej_jwriter "github.com/mailru/easyjson/jwriter"
)

// This conditionally-compiled file provides custom marshal/unmarshal functions for the User type
// in EasyJSON.
//
// EasyJSON's code generator does recognize the same MarshalJSON and UnmarshalJSON methods used by
// encoding/json, and will call them if present. But this mechanism is inefficient: when marshaling
// it requires the allocation of intermediate byte slices, and when unmarshaling it causes the
// JSON object to be parsed twice. It is preferable to have our marshal/unmarshal methods write to
// and read from the EasyJSON Writer/Lexer directly. Also, since user deserialization is a
// high-traffic path in some LaunchDarkly code on the service side, the extra overhead of the
// go-jsonstream abstraction is undesirable and we'll instead use an EasyJSON-generated
// deserializer for an intermediate struct type.
//
// For more information, see: https://gopkg.in/launchdarkly/go-jsonstream.v1

func (u User) MarshalEasyJSON(writer *ej_jwriter.Writer) {
	wrappedWriter := jwriter.NewWriterFromEasyJSONWriter(writer)
	u.WriteToJSONWriter(&wrappedWriter)
}

func (u *User) UnmarshalEasyJSON(lexer *jlexer.Lexer) {
	if lexer.IsNull() {
		lexer.Delim('{') // to trigger an "expected an object, got null" error
		return
	}
	var fields userEquivalentStruct
	fields.UnmarshalEasyJSON(lexer)
	if lexer.Error() != nil {
		return
	}
	if !fields.Key.IsDefined() {
		lexer.AddError(ErrMissingKey())
		return
	}
	u.key = fields.Key.StringValue()
	u.secondary = fields.Secondary
	u.ip = fields.Ip
	u.country = fields.Country
	u.email = fields.Email
	u.firstName = fields.FirstName
	u.lastName = fields.LastName
	u.avatar = fields.Avatar
	u.name = fields.Name
	u.anonymous = fields.Anonymous
	u.custom = fields.Custom
	if len(fields.PrivateAttributeNames) != 0 {
		u.privateAttributes = make(map[UserAttribute]struct{}, len(fields.PrivateAttributeNames))
		for _, a := range fields.PrivateAttributeNames {
			u.privateAttributes[UserAttribute(a)] = struct{}{}
		}
	}
}

//go:generate easyjson -build_tags launchdarkly_easyjson -output_filename user_serialization_easyjson_generated.go user_serialization_easyjson.go

//easyjson:json
type userEquivalentStruct struct {
	Key                   ldvalue.OptionalString `json:"key"`
	Secondary             ldvalue.OptionalString `json:"secondary,omitempty"`
	Ip                    ldvalue.OptionalString `json:"ip,omitempty"`
	Country               ldvalue.OptionalString `json:"country,omitempty"`
	Email                 ldvalue.OptionalString `json:"email,omitempty"`
	FirstName             ldvalue.OptionalString `json:"firstName,omitempty"`
	LastName              ldvalue.OptionalString `json:"lastName,omitempty"`
	Avatar                ldvalue.OptionalString `json:"avatar,omitempty"`
	Name                  ldvalue.OptionalString `json:"name,omitempty"`
	Anonymous             ldvalue.OptionalBool   `json:"anonymous,omitempty"`
	Custom                ldvalue.ValueMap       `json:"custom,omitempty"`
	PrivateAttributeNames []string               `json:"privateAttributeNames,omitempty"`
}
