package serdes

type StringSerializationProvider struct{}

func (s *StringSerializationProvider) InitSerializer(_ string, _ string) error {
	return nil
}

func (s *StringSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringSerializationProvider) Serialize(_ string, message any) ([]byte, error) {
	return []byte(message.(string)), nil
}

func (s *StringSerializationProvider) GetSchemaName() string {
	return ""
}

func (s *StringSerializationProvider) GetSchemaRegistryClient() any {
	return nil
}
