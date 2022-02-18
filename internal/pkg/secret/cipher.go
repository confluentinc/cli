package secret

type Cipher struct {
	Iterations       int
	KeyLength        int
	SaltDEK          string
	SaltMEK          string
	EncryptionAlgo   string
	EncryptedDataKey string
}

func NewCipher(cipherMode string) *Cipher {
	if cipherMode != "" {
		return &Cipher{
			Iterations:       MetadataKeyDefaultIterations,
			KeyLength:        MetadataKeyDefaultLengthBytes,
			SaltMEK:          "",
			SaltDEK:          "",
			EncryptionAlgo:   cipherMode,
			EncryptedDataKey: ""}
	} else {
		return &Cipher{
			Iterations:       MetadataKeyDefaultIterations,
			KeyLength:        MetadataKeyDefaultLengthBytes,
			SaltMEK:          "",
			SaltDEK:          "",
			EncryptionAlgo:   AES_CBC,
			EncryptedDataKey: ""}
	}
}
