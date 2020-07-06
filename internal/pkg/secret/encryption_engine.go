package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha512"
	"encoding/base64"
	"errors"
	"fmt"
	"math/rand"

	"golang.org/x/crypto/pbkdf2"

	"github.com/confluentinc/cli/internal/pkg/log"
)

// Encryption Engine performs Encryption, Decryption and Hash operations.
type EncryptionEngine interface {
	Encrypt(plainText string, key []byte) (string, string, error)
	Decrypt(cipher string, iv string, algo string, key []byte) (string, error)
	GenerateRandomDataKey(keyLength int) ([]byte, string, error)
	GenerateMasterKey(masterKeyPassphrase string, salt string) (string, string, error)
	WrapDataKey(dataKey []byte, masterKey string) (string, string, error)
	UnwrapDataKey(dataKey string, iv string, algo string, masterKey string) ([]byte, error)
}

// EncryptEngineImpl is the EncryptionEngine implementation
type EncryptEngineImpl struct {
	Cipher     *Cipher
	Logger     *log.Logger
	RandSource rand.Source
}

func NewEncryptionEngine(suite *Cipher, logger *log.Logger, randSource rand.Source) *EncryptEngineImpl {
	return &EncryptEngineImpl{Cipher: suite, Logger: logger, RandSource: randSource}
}

func (c *EncryptEngineImpl) generateRandomString(keyLength int) (string, error) {
	randomBytes := make([]byte, keyLength)
	r := rand.New(c.RandSource)
	_, err := r.Read(randomBytes)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(randomBytes), nil
}

func (c *EncryptEngineImpl) GenerateRandomDataKey(keyLength int) ([]byte, string, error) {

	// Generate random data key
	keyString, err := c.generateRandomString(keyLength)
	if err != nil {
		return []byte(""), "", err
	}

	// Generate random salt
	salt, err := c.generateRandomString(keyLength)
	if err != nil {
		return []byte(""), "", err
	}

	key, err := c.generateEncryptionKey(keyString, salt)
	if err != nil {
		return []byte(""), "", err
	}
	return key, salt, nil
}

func (c *EncryptEngineImpl) GenerateMasterKey(masterKeyPassphrase string, salt string) (string, string, error) {
	// Generate random salt
	var err error
	if salt == "" {
		salt, err = c.generateRandomString(MetadataKeyDefaultLengthBytes)
		if err != nil {
			return "", "", err
		}
	}

	key, err := c.generateEncryptionKey(masterKeyPassphrase, salt)
	if err != nil {
		return "", "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	return encodedKey, salt, nil
}

func (c *EncryptEngineImpl) WrapDataKey(dataKey []byte, masterKey string) (string, string, error) {
	dataKeyStr := base64.StdEncoding.EncodeToString(dataKey)
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return "", "", err
	}
	return c.Encrypt(dataKeyStr, masterKeyByte)
}

func (c *EncryptEngineImpl) UnwrapDataKey(dataKey string, iv string, _ string, masterKey string) ([]byte, error) {
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return []byte{}, err
	}

	dataKeyEnc, err := c.Decrypt(dataKey, iv, c.Cipher.EncryptionAlgo, masterKeyByte)
	if err != nil {
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(dataKeyEnc)
}

func (c *EncryptEngineImpl) Encrypt(plainText string, key []byte) (data string, ivStr string, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New("failed to encrypt the plain text:" + x)
			case error:
				err = x
			default:
				err = errors.New("failed to encrypt the plain text")
			}
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	ivStr, err = c.generateRandomString(aes.BlockSize)
	if err != nil {
		return "", "", err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(ivStr)
	if err != nil {
		return "", "", err
	}
	ecb := cipher.NewCBCEncrypter(block, ivBytes)
	content := []byte(plainText)
	content = c.pKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	result := base64.StdEncoding.EncodeToString(crypted)
	return result, ivStr, nil
}

func (c *EncryptEngineImpl) Decrypt(cipher string, iv string, _ string, key []byte) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}
	plainText, err := c.decrypt(cipherBytes, key, ivBytes)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func (c *EncryptEngineImpl) generateEncryptionKey(keyPhrase string, salt string) ([]byte, error) {
	key := pbkdf2.Key([]byte(keyPhrase), []byte(salt), c.Cipher.Iterations, c.Cipher.KeyLength, sha512.New)
	return key, nil
}

func (c *EncryptEngineImpl) decrypt(crypt []byte, key []byte, iv []byte) (plain []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New("failed to decrypt the cipher:" + x)
			case error:
				err = x
			default:
				err = errors.New("failed to decrypt the cipher")
			}
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	ecb := cipher.NewCBCDecrypter(block, iv)
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return c.pKCS5Trimming(decrypted)
}

func (c *EncryptEngineImpl) pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	length := len(ciphertext) % blockSize
	padding := blockSize - length
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (c *EncryptEngineImpl) pKCS5Trimming(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	length := len(encrypt) - int(padding)
	if length < 0 || length > len(encrypt) {
		return nil, fmt.Errorf("failed to decrypt the cipher: data is corrupted.")
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}
