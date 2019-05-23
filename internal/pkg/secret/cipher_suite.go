package secret

type Cipher interface {
	SetIterations(iterations int)
	SetKeyLength(length int)
	SetMasterKeyPath(path string)
	SetSalt(salt string)
}

type CipherSuite struct {
	Iterations       int
	KeyLength        int
	SaltDEK          string
	SaltMEK          string
	EncryptionAlgo   string
	EncryptedDataKey string
}

func NewCipherSuite(iterations int, keyLength int, masterKeyPath string, saltDEK string, saltMEK string, algo string, dataKey string) *CipherSuite {
	return &CipherSuite{Iterations: iterations, KeyLength: keyLength, SaltDEK: saltDEK, SaltMEK: saltMEK, EncryptionAlgo: algo, EncryptedDataKey: dataKey}
}

func NewDefaultCipherSuite() *CipherSuite {
	return &CipherSuite{
		Iterations:       METADATA_KEY_DEFAULT_ITERATIONS,
		KeyLength:        METADATA_KEY_DEFAULT_LENGTH_BYTES,
		SaltMEK:          METADATA_KEY_DEFAULT_SALT,
		SaltDEK:          "",
		EncryptionAlgo:   METADATA_ENC_ALGORITHM,
		EncryptedDataKey: ""}
}

func (c *CipherSuite) SetIterations(iterations int) {
	c.Iterations = iterations
}

func (c *CipherSuite) SetKeyLength(length int) {
	c.KeyLength = length
}

func (c *CipherSuite) SetSaltDEK(salt string) {
	c.SaltDEK = salt
}

func (c *CipherSuite) SetSaltMEK(salt string) {
	c.SaltDEK = salt
}

func (c *CipherSuite) SetDataKey(key string) {
	c.EncryptedDataKey = key
}
