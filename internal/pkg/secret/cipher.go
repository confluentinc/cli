package secret

type Cipher struct {
	Iterations       int
	KeyLength        int
	SaltDEK          string
	SaltMEK          string
	EncryptionAlgo   string
	EncryptedDataKey string
}

func NewDefaultCipher() *Cipher {
	return &Cipher{
		Iterations:       METADATA_KEY_DEFAULT_ITERATIONS,
		KeyLength:        METADATA_KEY_DEFAULT_LENGTH_BYTES,
		SaltMEK:          "",
		SaltDEK:          "",
		EncryptionAlgo:   METADATA_ENC_ALGORITHM,
		EncryptedDataKey: ""}
}
