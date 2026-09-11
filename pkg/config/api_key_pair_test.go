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
