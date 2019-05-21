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
	MasterKeyPath    string
	Salt             string
	EncryptionAlgo   string
	EncryptedDataKey string
}

func NewCipherSuite(iterations int, keyLength int, masterKeyPath string, salt string, algo string, dataKey string) *CipherSuite {
	return &CipherSuite{Iterations: iterations, KeyLength: keyLength, MasterKeyPath: masterKeyPath, Salt: salt, EncryptionAlgo: algo, EncryptedDataKey: dataKey}
}

func NewDefaultCipherSuite() *CipherSuite {
	return &CipherSuite{
		Iterations:       METADATA_KEY_DEFAULT_ITERATIONS,
		KeyLength:        METADATA_KEY_DEFAULT_LENGTH_BYTES,
		MasterKeyPath:    "",
		Salt:             METADATA_KEY_DEFAULT_SALT,
		EncryptionAlgo:   METADATA_ENC_ALGORITHM,
		EncryptedDataKey: ""}
}

func (c *CipherSuite) SetIterations(iterations int) {
	c.Iterations = iterations
}

func (c *CipherSuite) SetKeyLength(length int) {
	c.KeyLength = length
}

func (c *CipherSuite) SetMasterKeyPath(path string) {
	c.MasterKeyPath = path
}

func (c *CipherSuite) SetSalt(salt string) {
	c.Salt = salt
}

func (c *CipherSuite) SetDataKey(key string) {
	c.EncryptedDataKey = key
}
