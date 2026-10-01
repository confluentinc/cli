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

// An empty secret must never be handed to secret.Encrypt: DPAPI (CryptProtectData) rejects
// empty input with "The parameter is incorrect." on Windows, so an empty API secret must stay
// empty rather than be encrypted into ciphertext-of-"". No salt/nonce is derived either, so a
// credential that never had an API secret contributes nothing to the on-disk store.
func TestAPIKeyPair_EncryptSecret_LeavesEmptySecretEmpty(t *testing.T) {
	pair := &APIKeyPair{Key: "key", Secret: ""}

	require.NoError(t, pair.EncryptSecret())

	require.Empty(t, pair.Secret, "an empty secret must not be encrypted (DPAPI rejects empty input)")
	require.Nil(t, pair.Salt)
	require.Nil(t, pair.Nonce)
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
