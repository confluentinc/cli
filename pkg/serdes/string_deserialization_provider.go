package serdes

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) InitDeserializer(_ string, _ string) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(_ string, data []byte, message any) error {
	message = string(data)
	return nil
}
