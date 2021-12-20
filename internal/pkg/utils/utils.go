package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"time"

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

func Remove(haystack []string, needle string) []string {
	for i, x := range haystack {
		if x == needle {
			return append(haystack[:i], haystack[i+1:]...)
		}
	}
	return haystack
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

	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	data = NormalizeByteArrayNewLines(data)

	property, err := loader.LoadBytes(data)
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

func ValidateEmail(email string) bool {
	rgxEmail := regexp.MustCompile("^[a-zA-Z0-9.!#$%&'*+\\/=?^_`{|}~-]+@[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?(?:\\.[a-zA-Z0-9](?:[a-zA-Z0-9-]{0,61}[a-zA-Z0-9])?)*$")
	matched := rgxEmail.MatchString(email)
	return matched
}

func Abbreviate(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[0:maxLength] + "..."
}

func ToMap(configs []string) (map[string]string, error) {
	configMap := make(map[string]string)
	for _, cfg := range configs {
		pair := strings.SplitN(cfg, "=", 2)
		if len(pair) < 2 {
			return nil, fmt.Errorf(errors.FailedToParseConfigErrMsg, cfg)
		}
		configMap[pair[0]] = pair[1]
	}
	return configMap, nil
}

func ReadConfigsFromFile(configFile string) (map[string]string, error) {
	if configFile == "" {
		return map[string]string{}, nil
	}

	configContents, err := ioutil.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	// Create config map from the argument.
	var configs []string
	for _, s := range strings.Split(string(configContents), "\n") {
		// Filter out blank lines
		spaceTrimmed := strings.TrimSpace(s)
		if s != "" && spaceTrimmed[0] != '#' {
			configs = append(configs, spaceTrimmed)
		}
	}

	return ToMap(configs)
}

func CropString(s string, n int) string {
	const suffix = "..."
	if n-len(suffix) < len(s) {
		cropped := s[:n-len(suffix)] + suffix
		if len(cropped) < len(s) {
			return cropped
		}
	}
	return s
}

func FormatUnixTime(timeMs int64) string {
	time := time.Unix(0, timeMs * int64(time.Millisecond))
	return time.UTC().Format("2006-01-02 15:04:05 MST")
}
