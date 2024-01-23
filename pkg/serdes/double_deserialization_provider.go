package serdes

import (
	"encoding/binary"
	"fmt"
	"math"
)

type DoubleDeserializationProvider struct{}

func (DoubleDeserializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (DoubleDeserializationProvider) Deserialize(data []byte) (string, error) {
	if len(data) != 8 {
		return "", fmt.Errorf("the double key is invalid")
	}

	return fmt.Sprintf("%f", math.Float64frombits(binary.LittleEndian.Uint64(data))), nil
}
