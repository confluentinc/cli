//go:build windows

package secret

import (
	"github.com/billgraziano/dpapi"
)

func DeriveEncryptionKey(salt string) ([]byte, error) {
	return []byte{}, nil
}

func Encrypt(password, _ string) (string, error) {
	encryptedPassword, err := dpapi.Encrypt(password)
	if err != nil {
		return "", err
	}
	return encryptedPassword, nil
}

func Decrypt(encrypted, _ string) (string, error) {
	decryptedPassword, err := dpapi.Decrypt(encrypted)
	if err != nil {
		return "", err
	}
	return decryptedPassword, nil
}
