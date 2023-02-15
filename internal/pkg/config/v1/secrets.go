package v1

type Secrets struct {
	Salt  []byte `json:"salt,omitempty"`
	Nonce []byte `json:"nonce,omitempty"`
}
