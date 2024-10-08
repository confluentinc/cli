package serdes

import (
	"encoding/binary"
	"math"
	"strconv"
)

type DoubleSerializationProvider struct{}

func (DoubleSerializationProvider) InitSerializer(_ string, _ string) error {
	return nil
}

func (DoubleSerializationProvider) LoadSchema(_ string, _ map[string]string) error {
	return nil
}

func (DoubleSerializationProvider) Serialize(_ string, message any) ([]byte, error) {
	f, err := strconv.ParseFloat(message.(string), 64)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, 8)
	binary.LittleEndian.PutUint64(buf, math.Float64bits(f))

	return buf, nil
}

func (DoubleSerializationProvider) GetSchemaName() string {
	return ""
}
