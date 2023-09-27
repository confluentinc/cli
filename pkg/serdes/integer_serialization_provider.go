package serdes

import (
	"encoding/binary"
	"strconv"
)

type IntegerSerializationProvider struct{}

func (s *IntegerSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *IntegerSerializationProvider) Serialize(str string) ([]byte, error) {
	i, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(i))

	return buf, nil
}

func (s *IntegerSerializationProvider) GetSchemaName() string {
	return ""
}

func (s *IntegerSerializationProvider) SchemaBased() bool {
	return false
}
