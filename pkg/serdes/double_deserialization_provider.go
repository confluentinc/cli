package serdes

import (
	"encoding/binary"
	"fmt"
	"math"
)

type DoubleDeserializationProvider struct{}

func (DoubleDeserializationProvider) InitDeserializer(_ string, _ string, _ any) error {
	return nil
}

func (DoubleDeserializationProvider) Deserialize(_ string, data []byte) (string, error) {
	if len(data) == 0 {
		return "", nil
	}

	if len(data) != 8 {
		return "", fmt.Errorf("the double key is invalid")
	}

	message := fmt.Sprintf("%f", math.Float64frombits(binary.LittleEndian.Uint64(data)))
	return message, nil
}
