//go:build linux || darwin

package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"strconv"

	"github.com/denisbrodbeck/machineid"

	"golang.org/x/crypto/pbkdf2"
)

func DeriveEncryptionKey(salt string) ([]byte, error) {
	machineId, err := machineid.ID()
	if err != nil {
		return []byte{}, err
	}
	userId := strconv.Itoa(os.Getuid())
	encryptionKey := pbkdf2.Key([]byte(machineId+userId), []byte(salt), 10000, 32, sha256.New)
	return encryptionKey, nil
}

func Encrypt(password, salt string) (string, error) {
	encryptionKey, err := DeriveEncryptionKey(salt)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	encryptedPassword := aesgcm.Seal(nil, encryptionKey[:12], []byte(password), nil)

	return base64.RawStdEncoding.EncodeToString(encryptedPassword), nil
}

func Decrypt(encrypted, salt string) (string, error) {
	encryptionKey, err := DeriveEncryptionKey(salt)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(encryptionKey)
	if err != nil {
		return "", err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	cipherText, err := base64.RawStdEncoding.DecodeString(encrypted)
	if err != nil {
		return "", err
	}

	decryptedPassword, err := aesgcm.Open(nil, encryptionKey[:12], cipherText, nil)
	if err != nil {
		return "", err
	}

	return string(decryptedPassword), nil
}
