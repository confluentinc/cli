package config

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/confluentinc/cli/v4/pkg/secret"
)

// A concurrent session's merged-in secret already carries a platform prefix (AesGcm on Unix,
// Dpapi on Windows). EncryptSecret must leave it untouched rather than double-wrapping it.

func TestAPIKeyPair_EncryptSecret_SkipsAlreadyEncryptedDpapiSecret(t *testing.T) {
	dpapiSecret := secret.Dpapi + ":already-encrypted-on-windows"
	pair := &APIKeyPair{Key: "key", Secret: dpapiSecret}

	err := pair.EncryptSecret()

	require.NoError(t, err)
	require.Equal(t, dpapiSecret, pair.Secret)
}

func TestAPIKeyPair_EncryptSecret_SkipsAlreadyEncryptedAesGcmSecret(t *testing.T) {
	aesGcmSecret := secret.AesGcm + ":already-encrypted-on-unix"
	pair := &APIKeyPair{Key: "key", Secret: aesGcmSecret}

	err := pair.EncryptSecret()

	require.NoError(t, err)
	require.Equal(t, aesGcmSecret, pair.Secret)
}

// The cipher markers are stored as a "PREFIX:payload" pair, so the prefix guard
// must include the ":". A plaintext secret that merely begins with the marker
// word (no delimiter) must still be encrypted, not mistaken for ciphertext.
func TestAPIKeyPair_EncryptSecret_EncryptsPlaintextBeginningWithCipherWord(t *testing.T) {
	plain := secret.Dpapi + "-this-is-actually-plaintext"
	pair := &APIKeyPair{Key: "key", Secret: plain}

	require.NoError(t, pair.EncryptSecret())

	require.NotEqual(t, plain, pair.Secret, "a plaintext secret merely beginning with the cipher word must be encrypted")
	require.NoError(t, pair.DecryptSecret())
	require.Equal(t, plain, pair.Secret, "encryption of a marker-word plaintext must round-trip")
}
