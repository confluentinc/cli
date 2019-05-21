package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"github.com/confluentinc/cli/internal/pkg/log"
	"golang.org/x/crypto/pbkdf2"
	"io"
	_ "io/ioutil"
)

/**
* Encryption Engine performs Encryption, Decryption and Hash operations.
 */

type EncryptionEngine interface {
	AESEncrypt(plainText string, key []byte) (string, string, error)
	AESDecrypt(cipher string, iv string, algo string, key []byte) (string, error)
	GenerateRandomDataKey(keyLength int) ([]byte, error)
	GenerateMasterKey(masterKeyPassphrase string) (string, error)
	WrapDataKey(dataKey []byte, masterKey string) (string, string, error)
	UnWrapDataKey(dataKey string, iv string, algo string, masterKey string) ([]byte, error)
}

type EncryptEngineSuite struct {
	CipherSuite *CipherSuite
	Logger      *log.Logger
}

func NewEncryptionEngine(suite *CipherSuite, logger *log.Logger) *EncryptEngineSuite {
	return &EncryptEngineSuite{CipherSuite: suite, Logger: logger}
}

func (c *EncryptEngineSuite) GenerateRandomDataKey(keyLength int) ([]byte, error) {
	randomBytes := make([]byte, keyLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return randomBytes, err
	}

	randomString := base32.StdEncoding.EncodeToString(randomBytes)[:keyLength]
	key, err := c.generateEncryptionKey(randomString)
	if err != nil {
		return randomBytes, err
	}
	return key, nil
}

func (c *EncryptEngineSuite) GenerateMasterKey(masterKeyPassphrase string) (string, error) {
	key, err := c.generateEncryptionKey(masterKeyPassphrase)
	if err != nil {
		return "", err
	}

	encodedKey := base64.StdEncoding.EncodeToString(key)
	return encodedKey, nil
}

func (c *EncryptEngineSuite) WrapDataKey(dataKey []byte, masterKey string) (string, string, error) {
	dataKeyStr := base64.StdEncoding.EncodeToString(dataKey)
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return "", "", err
	}
	return c.AESEncrypt(dataKeyStr, masterKeyByte)
}

func (c *EncryptEngineSuite) UnWrapDataKey(dataKey string, iv string, algo string, masterKey string) ([]byte, error) {
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return []byte{}, err
	}

	dataKeyEnc, err := c.AESDecrypt(dataKey, iv, c.CipherSuite.EncryptionAlgo, masterKeyByte)
	if err != nil {
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(dataKeyEnc)
}

func (c *EncryptEngineSuite) AESEncrypt(plainText string, key []byte) (string, string, error) {
	return c.encryption(plainText, key)
}

func (c *EncryptEngineSuite) encryption(src string, key []byte) (string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		return "", "", err
	}
	ecb := cipher.NewCBCEncrypter(block, []byte(iv))
	content := []byte(src)
	content = c.pKCS5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	result := base64.StdEncoding.EncodeToString(crypted)
	ivStr := base64.StdEncoding.EncodeToString(iv)
	return result, ivStr, nil
}

func (c *EncryptEngineSuite) AESDecrypt(cipher string, iv string, algo string, key []byte) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}
	plainText, err := c.decryption(cipherBytes, key, ivBytes)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func (c *EncryptEngineSuite) generateEncryptionKey(keyPhrase string) ([]byte, error) {
	key := pbkdf2.Key([]byte(keyPhrase), []byte(c.CipherSuite.Salt), c.CipherSuite.Iterations, c.CipherSuite.KeyLength, sha512.New)
	return key, nil
}

func (c *EncryptEngineSuite) decryption(crypt []byte, key []byte, iv []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	ecb := cipher.NewCBCDecrypter(block, []byte(iv))
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return c.pKCS5Trimming(decrypted), nil
}

func (c *EncryptEngineSuite) pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	length := len(ciphertext) % blockSize
	padding := blockSize - length
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (c *EncryptEngineSuite) pKCS5Trimming(encrypt []byte) []byte {
	padding := encrypt[len(encrypt)-1]
	return encrypt[:len(encrypt)-int(padding)]
}

/* To do add HMAC functionality */
