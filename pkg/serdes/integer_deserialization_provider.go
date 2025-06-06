package serdes

import (
	"encoding/binary"
	"fmt"
)

type IntegerDeserializationProvider struct{}

func (IntegerDeserializationProvider) InitDeserializer(_, _, _ string, _ SchemaRegistryAuth, _ any) error {
	return nil
}

func (IntegerDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (IntegerDeserializationProvider) Deserialize(_ string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(data) != 4 {
		return "", fmt.Errorf("the integer key is invalid")
	}

	message := fmt.Sprintf("%d", binary.LittleEndian.Uint32(data))
	return message, nil
}
