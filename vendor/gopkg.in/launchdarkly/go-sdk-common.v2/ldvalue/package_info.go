// Package ldvalue provides an abstraction of the LaunchDarkly SDK's general value type. LaunchDarkly
// supports the standard JSON data types of null, boolean, number, string, array, and object (map), for
// any feature flag variation or custom user attribute. The ldvalue.Value type can contain any of these.
//
// This package also provides the helper types OptionalBool, OptionalInt, and OptionalString, which are
// safer alternatives to using pointers for optional values.
package ldvalue
