package serdes

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) InitDeserializer(_ string, _ string, _ any) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(_ string, data []byte) (string, error) {
	message := string(data)
	return message, nil
}
