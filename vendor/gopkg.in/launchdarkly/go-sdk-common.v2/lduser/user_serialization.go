package lduser

import (
	"encoding/json"

	"gopkg.in/launchdarkly/go-sdk-common.v2/jsonstream" //nolint:staticcheck // using a deprecated API
	"gopkg.in/launchdarkly/go-sdk-common.v2/ldvalue"

	"gopkg.in/launchdarkly/go-jsonstream.v1/jreader"
	"gopkg.in/launchdarkly/go-jsonstream.v1/jwriter"
)

var userRequiredProperties = []string{"key"} //nolint:gochecknoglobals

type missingKeyError struct{}

func (e missingKeyError) Error() string {
	return "User must have a key property"
}

// ErrMissingKey returns the standard error value that is used if you try to unmarshal a user from JSON
// and the "key" property is either absent or null. This is distinguished from other kinds of unmarshaling
// errors (such as trying to set a string property to a non-string value) in order to support use cases
// where incomplete data needs to be treated differently from malformed data.
//
// LaunchDarkly does allow a user to have an empty string ("") as a key in some cases, but this is
// discouraged since analytics events will not work properly without unique user keys.
func ErrMissingKey() error {
	return missingKeyError{}
}

// String returns a simple string representation of a user.
//
// This currently uses the same JSON string representation as User.MarshalJSON(). Do not rely on this
// specific behavior of String(); it is intended for convenience in debugging.
func (u User) String() string {
	bytes, _ := json.Marshal(u)
	return string(bytes)
}

// MarshalJSON provides JSON serialization for User when using json.MarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// user attributes are at the top level along with a "custom" property that is an object containing all of
// the custom attributes.
//
// In order for the representation to be as compact as possible, any top-level attributes for which no
// value has been set (as opposed to being set to an empty string) will be completely omitted, rather
// than including "attributeName":null in the JSON output. Similarly, if there are no custom attributes,
// there will be no "custom" property (rather than "custom":{}). This distinction does not matter to
// LaunchDarkly services-- they will treat an explicit null value in JSON data the same as an unset
// attribute, and treat an omitted "custom" the same as an empty "custom" map.
func (u User) MarshalJSON() ([]byte, error) {
	return jwriter.MarshalJSONWithWriter(u)
}

// UnmarshalJSON provides JSON deserialization for User when using json.UnmarshalJSON.
//
// This is LaunchDarkly's standard JSON representation for user properties, in which all of the built-in
// properties are at the top level along with a "custom" property that is an object containing all of
// the custom properties.
//
// Any property that is either completely omitted or has a null value is ignored and left in an unset
// state, except for "key". All users must have a key (even if it is ""), so an omitted or null "key"
// property causes the error ErrMissingKey().
//
// Trying to unmarshal any non-struct value, including a JSON null, into a User will return a
// json.UnmarshalTypeError. If you want to unmarshal optional user data that might be null, use *User
// instead of User.
func (u *User) UnmarshalJSON(data []byte) error {
	return jreader.UnmarshalJSONWithReader(data, u)
}

// ReadFromJSONReader provides JSON deserialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Unmarshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (u *User) ReadFromJSONReader(r *jreader.Reader) {
	var parsed User

	for obj := r.Object().WithRequiredProperties(userRequiredProperties); obj.Next(); {
		switch string(obj.Name()) {
		case "key":
			if key, nonNull := r.StringOrNull(); nonNull {
				parsed.key = key
			} else {
				r.AddError(ErrMissingKey())
			}
		case "anonymous":
			parsed.anonymous.ReadFromJSONReader(r)
		case "custom":
			parsed.custom.ReadFromJSONReader(r)
		case "privateAttributeNames":
			for arr := r.ArrayOrNull(); arr.Next(); {
				s := r.String()
				if r.Error() == nil {
					if parsed.privateAttributes == nil {
						parsed.privateAttributes = make(map[UserAttribute]struct{})
					}
					parsed.privateAttributes[UserAttribute(s)] = struct{}{}
				}
			}
		default:
			var setter func(*User, ldvalue.OptionalString)
			for _, sas := range optStringAttrSetters {
				if string(obj.Name()) == string(sas.attr) {
					setter = sas.setter
					break
				}
			}
			if setter != nil {
				var s ldvalue.OptionalString
				s.ReadFromJSONReader(r)
				if r.Error() == nil {
					setter(&parsed, s)
				}
			}
		}
	}
	if err := r.Error(); err != nil {
		if rpe, ok := err.(jreader.RequiredPropertyError); ok && rpe.Name == "key" {
			r.ReplaceError(ErrMissingKey())
		}
	} else {
		*u = parsed
	}
}

// WriteToJSONWriter provides JSON serialization for use with the jsonstream API.
//
// This implementation is used by the SDK in cases where it is more efficient than JSON.Marshal.
// See https://github.com/launchdarkly/go-jsonstream for more details.
func (u User) WriteToJSONWriter(w *jwriter.Writer) {
	obj := w.Object()
	obj.Name("key").String(u.key)
	optStringProperty(&obj, "secondary", u.secondary)
	optStringProperty(&obj, "ip", u.ip)
	optStringProperty(&obj, "country", u.country)
	optStringProperty(&obj, "email", u.email)
	optStringProperty(&obj, "firstName", u.firstName)
	optStringProperty(&obj, "lastName", u.lastName)
	optStringProperty(&obj, "avatar", u.avatar)
	optStringProperty(&obj, "name", u.name)
	obj.Maybe("anonymous", u.anonymous.IsDefined()).Bool(u.anonymous.BoolValue())
	if u.custom.Count() > 0 {
		u.custom.WriteToJSONWriter(obj.Name("custom"))
	}
	if len(u.privateAttributes) > 0 {
		arr := obj.Name("privateAttributeNames").Array()
		for name := range u.privateAttributes {
			arr.String(string(name))
		}
		arr.End()
	}
	obj.End()
}

func optStringProperty(obj *jwriter.ObjectState, name string, value ldvalue.OptionalString) {
	obj.Maybe(name, value.IsDefined()).String(value.StringValue())
}

// WriteToJSONBuffer provides JSON serialization for use with the deprecated jsonstream API.
//
// Deprecated: this method is provided for backward compatibility. The LaunchDarkly SDK no longer
// uses this API; instead it uses the newer https://github.com/launchdarkly/go-jsonstream.
func (u User) WriteToJSONBuffer(j *jsonstream.JSONBuffer) {
	jsonstream.WriteToJSONBufferThroughWriter(u, j)
}

type optStringAttrSetter struct {
	attr   UserAttribute
	setter func(*User, ldvalue.OptionalString)
}

//nolint:gochecknoglobals
var optStringAttrSetters = []optStringAttrSetter{
	{SecondaryKeyAttribute, func(u *User, s ldvalue.OptionalString) { u.secondary = s }},
	{IPAttribute, func(u *User, s ldvalue.OptionalString) { u.ip = s }},
	{CountryAttribute, func(u *User, s ldvalue.OptionalString) { u.country = s }},
	{EmailAttribute, func(u *User, s ldvalue.OptionalString) { u.email = s }},
	{FirstNameAttribute, func(u *User, s ldvalue.OptionalString) { u.firstName = s }},
	{LastNameAttribute, func(u *User, s ldvalue.OptionalString) { u.lastName = s }},
	{AvatarAttribute, func(u *User, s ldvalue.OptionalString) { u.avatar = s }},
	{NameAttribute, func(u *User, s ldvalue.OptionalString) { u.name = s }},
}
