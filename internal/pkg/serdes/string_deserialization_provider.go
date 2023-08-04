package serdes

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringDeserializationProvider) decode(data []byte) (string, error) {
	return string(data), nil
}
