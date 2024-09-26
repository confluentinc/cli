//go:build windows

package secret

import (
	"strings"

	"github.com/billgraziano/dpapi"
	"github.com/confluentinc/cli/v4/pkg/log"
)

func generateRandomBytes(_ int) ([]byte, error) {
	return nil, nil
}

func GenerateSaltAndNonce() ([]byte, []byte, error) {
	return nil, nil, nil
}

func DeriveEncryptionKey(_ string) ([]byte, error) {
	return nil, nil
}

func Encrypt(_, password string, _, _ []byte) (string, error) {
	encryptedPassword, err := dpapi.Encrypt(password)
	if err != nil {
		return "", err
	}
	return Dpapi + ":" + encryptedPassword, nil
}

func Decrypt(_, encrypted string, _, _ []byte) (string, error) {
	log.CliLogger.Tracef("Decrypting secret: %s", encrypted)
	encrypted = strings.TrimPrefix(encrypted, Dpapi+":")
	decryptedPassword, err := dpapi.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return decryptedPassword, nil
}
