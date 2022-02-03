package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"io"

	"golang.org/x/crypto/pbkdf2"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

// Encryption Engine performs Encryption, Decryption and Hash operations.
type EncryptionEngine interface {
	Encrypt(plainText string, key []byte) (string, string, error)
	Decrypt(cipher string, nonce string, algo string, key []byte) (string, error)
	GenerateRandomDataKey(keyLength int) ([]byte, string, error)
	GenerateMasterKey(masterKeyPassphrase string, salt string) (string, string, error)
	WrapDataKey(dataKey []byte, masterKey string) (string, string, error)
	UnwrapDataKey(dataKey string, nonce string, algo string, masterKey string) ([]byte, error)
}

// EncryptEngineImpl is the EncryptionEngine implementation
type EncryptEngineImpl struct {
	Cipher *Cipher
	Logger *log.Logger
}

func NewEncryptionEngine(suite *Cipher) *EncryptEngineImpl {
	return &EncryptEngineImpl{Cipher: suite}
}

func (c *EncryptEngineImpl) generateRandomString(keyLength int) (string, error) {
	randomBytes := make([]byte, keyLength)
	_, err := rand.Read(randomBytes)
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

func (c *EncryptEngineImpl) UnwrapDataKey(dataKey string, nonce string, _ string, masterKey string) ([]byte, error) {
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return []byte{}, err
	}

	dataKeyEnc, err := c.Decrypt(dataKey, nonce, c.Cipher.EncryptionAlgo, masterKeyByte)
	if err != nil {
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(dataKeyEnc)
}

func (c *EncryptEngineImpl) Encrypt(plainText string, key []byte) (data string, nonceStr string, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(errors.EncryptPlainTextErrorMsg + ": " + x)
			case error:
				err = x
			default:
				err = errors.New(errors.EncryptPlainTextErrorMsg)
			}
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	nonceBytes := make([]byte, 12)
	if _, err := io.ReadFull(rand.Reader, nonceBytes); err != nil {
		panic(err.Error())
	}

	nonceStr = base64.StdEncoding.EncodeToString(nonceBytes)

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	content := []byte(plainText)
	content = c.pKCS5Padding(content, block.BlockSize())

	ciphertext := aesgcm.Seal(nil, nonceBytes, content, nil)
	result := base64.StdEncoding.EncodeToString(ciphertext)
	return result, nonceStr, nil
}

func (c *EncryptEngineImpl) Decrypt(cipher string, nonce string, _ string, key []byte) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	nonceBytes, err := base64.StdEncoding.DecodeString(nonce)
	if err != nil {
		return "", err
	}
	plainText, err := c.decrypt(cipherBytes, key, nonceBytes)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func (c *EncryptEngineImpl) generateEncryptionKey(keyPhrase string, salt string) ([]byte, error) {
	key := pbkdf2.Key([]byte(keyPhrase), []byte(salt), c.Cipher.Iterations, c.Cipher.KeyLength, sha512.New)
	return key, nil
}

func (c *EncryptEngineImpl) decrypt(crypt []byte, key []byte, nonce []byte) (plain []byte, err error) {
	defer func() {
		if r := recover(); r != nil {
			switch x := r.(type) {
			case string:
				err = errors.New(errors.DecryptCypherErrorMsg + ": " + x)
			case error:
				err = x
			default:
				err = errors.New(errors.DecryptCypherErrorMsg)
			}
		}
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	decrypted, err := aesgcm.Open(nil, nonce, crypt, nil)
	if err != nil {
		panic(err.Error())
	}

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
		return nil, errors.New(errors.DataCorruptedErrorMsg)
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}
