package serdes

type StringSerializationProvider struct{}

func (s *StringSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *StringSerializationProvider) Serialize(str string) ([]byte, error) {
	return []byte(str), nil
}

func (s *StringSerializationProvider) GetSchemaName() string {
	return ""
}

func (s *StringSerializationProvider) SchemaBased() bool {
	return false
}
