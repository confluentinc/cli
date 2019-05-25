package secret

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
