package secret

import (
	"regexp"
	"strings"
)

func GenerateConfigValue(key string, path string) string {
	return "${securePass:" + path + ":" + key + "}"
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
