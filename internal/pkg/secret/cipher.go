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
	
	cipher := &Cipher{
		Iterations:     MetadataKeyDefaultIterations,
		KeyLength:      MetadataKeyDefaultLengthBytes,
		EncryptionAlgo: AES_CBC,
	}
	
	if cipherMode != "" {
		cipher.EncryptionAlgo = cipherMode
	}
	
	return cipher
}
