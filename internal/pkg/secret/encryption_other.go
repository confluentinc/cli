//go:build linux || darwin

package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/panta/machineid"
	"golang.org/x/crypto/pbkdf2"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

const (
	iterationNumber = 10000
	keyLength       = 32
)

func GenerateRandomBytes(n int) ([]byte, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return []byte{}, err
	}
	return b, nil
}

func DeriveEncryptionKey(salt []byte) ([]byte, error) {
	machineId, err := machineid.ID()
	if err != nil {
		return []byte{}, err
	}
	pwd := []byte(fmt.Sprintf("%s%d", machineId, os.Getuid()))
	return pbkdf2.Key(pwd, salt, iterationNumber, keyLength, sha256.New), nil
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

	if len(nonce) != NonceLength {
		return "", errors.New(errors.IncorrectNonceLengthErrorMsg)
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
		return "", fmt.Errorf("failed to decode base64: %v", err)
	}

	if len(nonce) != NonceLength {
		return "", errors.New(errors.IncorrectNonceLengthErrorMsg)
	}
	decryptedPassword, err := aesgcm.Open(nil, nonce, cipherText, []byte(username))
	if err != nil {
		return "", fmt.Errorf("CLI does not have write permission for `/etc/machine-id`, or `~/.confluent/config.json` is corrupted: %w", err)
	}

	return string(decryptedPassword), nil
}
