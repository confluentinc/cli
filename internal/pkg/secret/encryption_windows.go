//go:build windows

package secret

import (
	"github.com/billgraziano/dpapi"
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
	return encryptedPassword, nil
}

func Decrypt(_, encrypted string, _, _ []byte) (string, error) {
	decryptedPassword, err := dpapi.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return decryptedPassword, nil
}
