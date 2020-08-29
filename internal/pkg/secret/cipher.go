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
		Iterations:       MetadataKeyDefaultIterations,
		KeyLength:        MetadataKeyDefaultLengthBytes,
		SaltMEK:          "",
		SaltDEK:          "",
		EncryptionAlgo:   MetadataEncAlgorithm,
		EncryptedDataKey: ""}
}
