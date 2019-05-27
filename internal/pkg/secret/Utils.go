package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"github.com/magiconair/properties"
)

func GenerateConfigValue(key string, path string) string {
	return "${securepass:" + path + ":" + key + "}"
}

func ParseCipherValue(cipher string) (string, string, string) {
	data := findMatchTrim(cipher, "data\\:(.*?)\\,", "data:", ",")
	iv := findMatchTrim(cipher, "iv\\:(.*?)\\,", "iv:", ",")
	algo := findMatchTrim(cipher, "ENC\\[(.*?)\\,", "ENC[", ",")
	return data, iv, algo
}

func findMatchTrim(original string, pattern string, prefix string, suffix string) string {
	re := regexp.MustCompile(pattern)
	match := re.FindStringSubmatch(original)
	substring := ""
	if len(match) != 0 {
		substring = strings.TrimPrefix(strings.TrimSuffix(match[0], suffix), prefix)
	}
	return substring
}

func WritePropertiesFile(path string, property *properties.Properties) error {
	buf := new(bytes.Buffer)
	_, err := property.WriteComment(buf, "# ", properties.ISO_8859_1)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, buf.Bytes(), 0644)
	return err
}

func IsPathValid(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !IsPathValid(path) {
		return nil, fmt.Errorf("Invalid file path.")
	}
	property := properties.MustLoadFile(path, properties.ISO_8859_1)
	return property, nil
}
