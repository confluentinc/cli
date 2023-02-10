package v1

type LoginCredential struct {
	IsCloud           bool   `json:"is_cloud"`
	Url               string `json:"url"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"encrypted_password"`
	Salt              []byte `json:"salt,omitempty"`
	Nonce             []byte `json:"nonce,omitempty"`
}
