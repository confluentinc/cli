package utils

import (
	"github.com/confluentinc/cli/internal/pkg/errors"
	"github.com/confluentinc/properties"
	"os"
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
	property, err := loader.LoadFile(path)
	if err != nil {
		return nil, err
	}
	property.DisableExpansion = true
	return property, nil
}
