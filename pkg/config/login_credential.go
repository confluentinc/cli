package config

type LoginCredential struct {
	IsCloud           bool   `json:"is_cloud"`
	Url               string `json:"url"`
	Username          string `json:"username"`
	EncryptedPassword string `json:"-"`
	Salt              []byte `json:"-"`
	Nonce             []byte `json:"-"`
}
