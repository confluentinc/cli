package serdes

type StringDeserializationProvider struct{}

func (s *StringDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ any) error {
	return nil
}

func (s *StringDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringDeserializationProvider) Deserialize(_ string, data []byte) (string, error) {
	message := string(data)
	return message, nil
}
