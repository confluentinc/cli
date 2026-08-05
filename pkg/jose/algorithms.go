// Package jose exposes the signature-algorithm allowlist shared by every
// JWT parser in the CLI. go-jose/v4 requires callers to pass an allowlist
// to ParseSigned, and centralizing it here keeps pkg/jwt and pkg/config
// from drifting on which token shapes they will parse.
package jose

import jose "github.com/go-jose/go-jose/v4"

// SignatureAlgorithms is intentionally broad: callers read claims via
// UnsafeClaimsWithoutVerification (no signature check), so the allowlist
// is required by go-jose/v4 to parse but does not gate any security
// decision in the CLI.
var SignatureAlgorithms = []jose.SignatureAlgorithm{
	jose.RS256, jose.RS384, jose.RS512,
	jose.ES256, jose.ES384, jose.ES512,
	jose.PS256, jose.PS384, jose.PS512,
	jose.HS256, jose.HS384, jose.HS512,
	jose.EdDSA,
}
