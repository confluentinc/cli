package utils

import (
	"bytes"
	"io/ioutil"
	"os"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/internal/pkg/errors"
)

func Max(x, y int64) int64 {
	if x > y {
		return x
	}
	return y
}

func TestEq(a, b []string) bool {
	// If one is nil, the other must also be nil.
	if (a == nil) != (b == nil) {
		return false
	}

	if len(a) != len(b) {
		return false
	}

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}

func RemoveDuplicates(s []string) []string {
	check := make(map[string]int)
	for _, v := range s {
		check[v] = 0
	}
	var noDups []string
	for k := range check {
		noDups = append(noDups, k)
	}
	return noDups
}

func Contains(haystack []string, needle string) bool {
	for _, x := range haystack {
		if x == needle {
			return true
		}
	}
	return false
}

func DoesPathExist(path string) bool {
	if path == "" {
		return false
	}

	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, errors.Errorf(errors.InvalidFilePathErrorMsg, path)
	}
	loader := new(properties.Loader)
	loader.Encoding = properties.UTF8
	loader.PreserveFormatting = true
	//property.DisableExpansion = true
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	bytes = NormalizeByteArrayNewLines(bytes)
	property, err := loader.LoadBytes(bytes)
	if err != nil {
		return nil, err
	}
	property.DisableExpansion = true
	return property, nil
}

// NormalizeNewLines replaces \r\n and \r newline sequences with \n
func NormalizeNewLines(raw string) string {
	return string(NormalizeByteArrayNewLines([]byte(raw)))
}

func NormalizeByteArrayNewLines(raw []byte) []byte {
	normalized := bytes.Replace(raw, []byte{13, 10}, []byte{10}, -1)
	normalized = bytes.Replace(normalized, []byte{13}, []byte{10}, -1)
	return normalized
}
