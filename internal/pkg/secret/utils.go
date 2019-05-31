package secret

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/magiconair/properties"
)

var dataRegex = regexp.MustCompile(DATA_PATTERN)
var ivRegex = regexp.MustCompile(IV_PATTERN)
var algoRegex = regexp.MustCompile(ENC_PATTERN)
var passwordRegex = regexp.MustCompile(PASSWORD_PATTERN)
var cipherRegex = regexp.MustCompile(CIPHER_PATTERN)

func GenerateConfigValue(key string, path string) string {
	return "${securepass:" + path + ":" + key + "}"
}

func ParseCipherValue(cipher string) (string, string, string) {
	data := findMatchTrim(cipher, dataRegex, "data:", ",")
	iv := findMatchTrim(cipher, ivRegex, "iv:", ",")
	algo := findMatchTrim(cipher, algoRegex, "ENC[", ",")
	return data, iv, algo
}

func findMatchTrim(original string, re *regexp.Regexp, prefix string, suffix string) string {
	match := re.FindStringSubmatch(original)
	substring := ""
	if len(match) != 0 {
		substring = strings.TrimPrefix(strings.TrimSuffix(match[0], suffix), prefix)
	}
	return substring
}

func WritePropertiesFile(path string, property *properties.Properties, writeComments bool) error {
	buf := new(bytes.Buffer)
	if writeComments {
		_, err := property.WriteComment(buf, "# ", properties.UTF8)
		if err != nil {
			return err
		}
	} else {
		_, err := property.Write(buf, properties.UTF8)
		if err != nil {
			return err
		}

	}

	err := ioutil.WriteFile(path, buf.Bytes(), 0644)
	return err
}

func DoesPathExist(path string) bool {
	if path == "" {
		return false
	}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return false
	}
	return true
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, fmt.Errorf("Invalid file path.")
	}
	property, err := properties.LoadFile(path, properties.UTF8)
	if err != nil {
		return nil, err
	}
	property.DisableExpansion = true
	return property, nil
}

func GenerateConfigKey(path string, key string) string {
	fileName := filepath.Base(path)
	// Intentionally not using the filepath.Join(fileName, key), because even if this CLI is run on Windows we know that
	// the server-side version will be running on a *nix variant and will thus have forward slashes to lookup the correct path
	return fileName + "/" + key
}
