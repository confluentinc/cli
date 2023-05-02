package secret

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
	"golang.org/x/crypto/pbkdf2"
)

const (
	SaltLength  = 24
	NonceLength = 12
)

// Encryption Engine performs Encryption, Decryption and Hash operations.
type EncryptionEngine interface {
	Encrypt(plainText string, key []byte, algo string) (string, string, error)
	Decrypt(cipher string, iv string, algo string, key []byte) (string, error)
	GenerateRandomDataKey(keyLength int) ([]byte, string, error)
	GenerateMasterKey(masterKeyPassphrase string, salt string) (string, string, error)
	WrapDataKey(dataKey []byte, masterKey string, algo string) (string, string, error)
	UnwrapDataKey(dataKey string, iv string, algo string, masterKey string) ([]byte, error)
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

func (c *EncryptEngineImpl) WrapDataKey(dataKey []byte, masterKey string, algo string) (string, string, error) {
	dataKeyStr := base64.StdEncoding.EncodeToString(dataKey)
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return "", "", err
	}
	return c.Encrypt(dataKeyStr, masterKeyByte, algo)
}

func (c *EncryptEngineImpl) UnwrapDataKey(dataKey string, iv string, algo string, masterKey string) ([]byte, error) {
	masterKeyByte, err := base64.StdEncoding.DecodeString(masterKey)
	if err != nil {
		return []byte{}, err
	}

	dataKeyEnc, err := c.Decrypt(dataKey, iv, algo, masterKeyByte)
	if err != nil {
		return []byte{}, err
	}

	return base64.StdEncoding.DecodeString(dataKeyEnc)
}

func (c *EncryptEngineImpl) Encrypt(plainText string, key []byte, algo string) (data string, ivStr string, err error) {
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

	if algo == AES_CBC { // Backwards compatability
		return c.encryptCBCMode(plainText, key)
	} else {
		return c.encryptGCMMode(plainText, key)
	}
}

func (c *EncryptEngineImpl) Decrypt(cipher string, iv string, algo string, key []byte) (string, error) {
	cipherBytes, err := base64.StdEncoding.DecodeString(cipher)
	if err != nil {
		return "", err
	}
	ivBytes, err := base64.StdEncoding.DecodeString(iv)
	if err != nil {
		return "", err
	}
	plainText, err := c.decrypt(cipherBytes, key, ivBytes, algo)
	if err != nil {
		return "", err
	}

	return string(plainText), nil
}

func (c *EncryptEngineImpl) generateEncryptionKey(keyPhrase string, salt string) ([]byte, error) {
	key := pbkdf2.Key([]byte(keyPhrase), []byte(salt), c.Cipher.Iterations, c.Cipher.KeyLength, sha512.New)
	return key, nil
}

func (c *EncryptEngineImpl) encryptCBCMode(plainText string, key []byte) (data string, ivStr string, err error) {
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
	content = c.pkcs5Padding(content, block.BlockSize())
	crypted := make([]byte, len(content))
	ecb.CryptBlocks(crypted, content)
	result := base64.StdEncoding.EncodeToString(crypted)
	return result, ivStr, nil
}

func (c *EncryptEngineImpl) encryptGCMMode(plainText string, key []byte) (string, string, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return "", "", err
	}

	ivStr, err := c.generateRandomString(MetadataIVLength)
	if err != nil {
		return "", "", err
	}

	ivBytes, err := base64.StdEncoding.DecodeString(ivStr)
	if err != nil {
		return "", "", err
	}

	aesGcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", "", err
	}

	content := []byte(plainText)

	ciphertext := aesGcm.Seal(nil, ivBytes, content, nil)
	result := base64.StdEncoding.EncodeToString(ciphertext)
	return result, ivStr, nil
}

func (c *EncryptEngineImpl) decrypt(crypt []byte, key []byte, iv []byte, algo string) (plain []byte, err error) {
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

	var decrypted []byte

	if algo == AES_CBC { // Backwards compatability
		ecb := cipher.NewCBCDecrypter(block, iv)
		decrypted = make([]byte, len(crypt))
		ecb.CryptBlocks(decrypted, crypt)
		return c.pkcs5Trimming(decrypted)
	} else if algo == AES_GCM {
		aesGcm, err := cipher.NewGCM(block)
		if err != nil {
			return []byte{}, err
		}

		decrypted, err = aesGcm.Open(nil, iv, crypt, nil)
		if err != nil {
			return []byte{}, err
		}
		return decrypted, nil
	} else {
		return []byte{}, errors.Errorf(errors.InvalidAlgorithmErrorMsg, algo)
	}
}

func (c *EncryptEngineImpl) pkcs5Trimming(encrypt []byte) ([]byte, error) {
	padding := encrypt[len(encrypt)-1]
	length := len(encrypt) - int(padding)
	if length < 0 || length > len(encrypt) {
		return nil, errors.New(errors.DataCorruptedErrorMsg)
	}
	return encrypt[:len(encrypt)-int(padding)], nil
}

func (c *EncryptEngineImpl) pkcs5Padding(ciphertext []byte, blockSize int) []byte {
	length := len(ciphertext) % blockSize
	padding := blockSize - length
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
