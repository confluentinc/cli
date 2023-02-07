package v1

type LoginCredential struct {
	IsCloud    bool   `json:"is_cloud"`
	IgnoreCert bool   `json:"ignore_cert"`
	Url        string `json:"url"`
	Username   string `json:"username"`
	Password   string `json:"password"`
	Salt       []byte `json:"salt,omitempty"`
	Nonce      []byte `json:"nonce,omitempty"`
}
