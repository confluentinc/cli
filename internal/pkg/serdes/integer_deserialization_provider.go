package serdes

import (
	"encoding/binary"
	"fmt"
)

type IntegerDeserializationProvider struct{}

func (s *IntegerDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (s *IntegerDeserializationProvider) decode(data []byte) (string, error) {
	return fmt.Sprintf("%d", binary.BigEndian.Uint32(data)), nil
}
