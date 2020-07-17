package serdes

type StringSerializationProvider uint32

func (stringProvider *StringSerializationProvider) getSchemaName() string {
	return "STRING"
}

func (stringProvider *StringSerializationProvider) encode(str string, schemaPath string) ([]byte, error) {
	// Simply returns bytes in string.
	return []byte(str), nil
}

type StringDeserializationProvider uint32

func (stringProvider *StringDeserializationProvider) getSchemaName() string {
	return "STRING"
}

func (stringProvider *StringDeserializationProvider) decode(data []byte, schemaPath string) (string, error) {
	// Simply wraps up bytes in string and returns.
	return string(data), nil
}
