//go:build boringcrypto

package fips

import _ "crypto/tls/fipsonly" // including this package when boringcrypto is enabled forces tls to use fips settings only
