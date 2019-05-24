package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base32"
	"encoding/base64"
	"errors"
	"fmt"
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
	GenerateRandomDataKey(keyLength int) ([]byte, string, error)
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

func (c *EncryptEngineSuite) generateRandomString(keyLength int) (string, error) {
	randomBytes := make([]byte, keyLength)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	randomString := base32.StdEncoding.EncodeToString(randomBytes)[:keyLength]
	return randomString, nil
}

func (c *EncryptEngineSuite) GenerateRandomDataKey(keyLength int) ([]byte, string, error) {

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

func (c *EncryptEngineSuite) GenerateMasterKey(masterKeyPassphrase string) (string, error) {
	key, err := c.generateEncryptionKey(masterKeyPassphrase, c.CipherSuite.SaltMEK)
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
	return c.encrypt(plainText, key)
}

func (c *EncryptEngineSuite) encrypt(src string, key []byte) (data string, ivStr string, err error) {
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
	ivStr = base64.StdEncoding.EncodeToString(iv)
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
	plainText, err := c.decrypt(cipherBytes, key, ivBytes)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func (c *EncryptEngineSuite) generateEncryptionKey(keyPhrase string, salt string) ([]byte, error) {
	key := pbkdf2.Key([]byte(keyPhrase), []byte(salt), c.CipherSuite.Iterations, c.CipherSuite.KeyLength, sha512.New)
	return key, nil
}

func (c *EncryptEngineSuite) decrypt(crypt []byte, key []byte, iv []byte) (plain []byte, err error) {
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

	ecb := cipher.NewCBCDecrypter(block, []byte(iv))
	decrypted := make([]byte, len(crypt))
	ecb.CryptBlocks(decrypted, crypt)

	return c.pKCS5Trimming(decrypted)
}

func (c *EncryptEngineSuite) pKCS5Padding(ciphertext []byte, blockSize int) []byte {
	length := len(ciphertext) % blockSize
	padding := blockSize - length
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func (c *EncryptEngineSuite) pKCS5Trimming(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	length := len(encrypt)-int(padding)
	if length < 0 || length > len(encrypt) {
		return nil, fmt.Errorf("failed to decrypt the cipher: data is corrupted." )
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}

/* To do add HMAC functionality */
