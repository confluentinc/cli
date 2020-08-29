package serdes

type RawSerializationProvider struct{}

func (rawProvider *RawSerializationProvider) LoadSchema(_ string) error {
	return nil
}

func (rawProvider *RawSerializationProvider) GetSchemaName() string {
	return RAWSCHEMANAME
}

func (rawProvider *RawSerializationProvider) encode(str string) ([]byte, error) {
	// Simply returns bytes in string.
	return []byte(str), nil
}

type RawDeserializationProvider struct{}

func (rawProvider *RawDeserializationProvider) LoadSchema(_ string) error {
	return nil
}

func (rawProvider *RawDeserializationProvider) GetSchemaName() string {
	return RAWSCHEMANAME
}

func (rawProvider *RawDeserializationProvider) decode(data []byte) (string, error) {
	// Simply wraps up bytes in string and returns.
	return string(data), nil
}
