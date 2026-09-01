package iam

// authMethodFormats maps the iam.v2 AUTH_TYPE_* enum values the API returns to the
// human-friendly labels the CLI has always displayed. Hand-written: the OpenAPI spec
// only declares the raw enum, so the generator cannot produce this mapping.
// TODO: discuss with Identity Team on the response
var authMethodFormats = map[string]string{
	"AUTH_TYPE_LOCAL":   "Username/Password",
	"AUTH_TYPE_SSO":     "SSO",
	"AUTH_TYPE_UNKNOWN": "Unknown",
}
