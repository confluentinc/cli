package secret

type Cipher struct {
	Iterations       int
	KeyLength        int
	SaltDEK          string
	SaltMEK          string
	EncryptionAlgo   string
	EncryptedDataKey string
}

func NewCipher() *Cipher {

	return &Cipher{
		Iterations:     MetadataKeyDefaultIterations,
		KeyLength:      MetadataKeyDefaultLengthBytes,
		EncryptionAlgo: AesGcm,
	}
}
