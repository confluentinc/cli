//go:build linux || darwin

package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"os"
	"strconv"

	"github.com/panta/machineid"

	"golang.org/x/crypto/pbkdf2"
)

const (
	iterationNumber = 10000
	keyLength       = 32
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return []byte{}, err
	}
	return b, err
}

func DeriveEncryptionKey(salt []byte) ([]byte, error) {
	machineId, err := machineid.ID()
	if err != nil {
		return []byte{}, err
	}
	userId := strconv.Itoa(os.Getuid())
	encryptionKey := pbkdf2.Key([]byte(machineId+userId), salt, iterationNumber, keyLength, sha256.New)
	return encryptionKey, nil
}

func Encrypt(username, password string, salt, nonce []byte) (string, error) {
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

	encryptedPassword := aesgcm.Seal(nil, nonce, []byte(password), []byte(username))

	return base64.RawStdEncoding.EncodeToString(encryptedPassword), nil
}

func Decrypt(username, encrypted string, salt, nonce []byte) (string, error) {
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

	decryptedPassword, err := aesgcm.Open(nil, nonce, cipherText, []byte(username))
	if err != nil {
		return "", err
	}

	return string(decryptedPassword), nil
}
