package types

type ApplicationOptions struct {
	DefaultProperties map[string]string
	FlinkGatewayUrl   string
	UnsafeTrace       bool
	UserAgent         string
}
