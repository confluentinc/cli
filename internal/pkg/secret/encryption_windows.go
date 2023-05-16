//go:build windows
// +build windows

package secret

import (
	"github.com/billgraziano/dpapi"
)

func GenerateRandomBytes(n int) ([]byte, error) {
	return []byte{}, nil
}

func DeriveEncryptionKey(salt string) ([]byte, error) {
	return []byte{}, nil
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
