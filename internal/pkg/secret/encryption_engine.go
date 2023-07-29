package secret

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha512"
	"encoding/base64"

	"golang.org/x/crypto/pbkdf2"

	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/cli/internal/pkg/log"
)

const (
	SaltLength  = 24
	NonceLength = 12
)

// Encryption Engine performs Encryption, Decryption and Hash operations.
type EncryptionEngine interface {
	Encrypt(plainText string, key []byte) (string, string, error)
	Decrypt(cipher, iv, algo string, key []byte) (string, error)
	GenerateRandomDataKey(keyLength int) ([]byte, string, error)
	GenerateMasterKey(masterKeyPassphrase, salt string) (string, string, error)
	WrapDataKey(dataKey []byte, masterKey string) (string, string, error)
	UnwrapDataKey(dataKey, iv, algo, masterKey string) ([]byte, error)
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
	if _, err := rand.Read(randomBytes); err != nil {
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

	key := c.generateEncryptionKey(keyString, salt)

	return key, salt, nil
}

func (c *EncryptEngineImpl) GenerateMasterKey(masterKeyPassphrase, salt string) (string, string, error) {
	// Generate random salt
	var err error
	if salt == "" {
		salt, err = c.generateRandomString(MetadataKeyDefaultLengthBytes)
		if err != nil {
			return "", "", err
		}
	}

	key := c.generateEncryptionKey(masterKeyPassphrase, salt)
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

func (c *EncryptEngineImpl) UnwrapDataKey(dataKey, iv, algo, masterKey string) ([]byte, error) {
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

func (c *EncryptEngineImpl) Encrypt(plainText string, key []byte) (string, string, error) {
	defer func() {
		_ = recover()
	}()

	return c.encrypt(plainText, key)
}

func (c *EncryptEngineImpl) Decrypt(cipher, iv, algo string, key []byte) (string, error) {
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

func (c *EncryptEngineImpl) generateEncryptionKey(keyPhrase, salt string) []byte {
	return pbkdf2.Key([]byte(keyPhrase), []byte(salt), c.Cipher.Iterations, c.Cipher.KeyLength, sha512.New)
}

func (c *EncryptEngineImpl) encrypt(plainText string, key []byte) (string, string, error) {
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

func (c *EncryptEngineImpl) decrypt(crypt, key, iv []byte, algo string) ([]byte, error) {
	defer func() {
		_ = recover()
	}()

	block, err := aes.NewCipher(key)
	if err != nil {
		return []byte{}, err
	}

	var decrypted []byte

	if algo == AesCbc { // Backwards compatibility
		ecb := cipher.NewCBCDecrypter(block, iv)
		decrypted = make([]byte, len(crypt))
		ecb.CryptBlocks(decrypted, crypt)
		return c.pkcs5Trimming(decrypted)
	} else if algo == AesGcm {
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
