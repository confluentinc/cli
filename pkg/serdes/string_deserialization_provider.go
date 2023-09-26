package serdes

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(data []byte) (string, error) {
	return string(data), nil
}
