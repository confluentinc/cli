package serdes

type RawSerializationProvider uint32

func (rawProvider *RawSerializationProvider) GetSchemaName() string {
	return "RAW"
}

func (rawProvider *RawSerializationProvider) encode(str string, schemaPath string) ([]byte, error) {
	// Simply returns bytes in string.
	return []byte(str), nil
}

type RawDeserializationProvider uint32

func (rawProvider *RawDeserializationProvider) GetSchemaName() string {
	return "RAW"
}

func (rawProvider *RawDeserializationProvider) decode(data []byte, schemaPath string) (string, error) {
	// Simply wraps up bytes in string and returns.
	return string(data), nil
}
