package utils

import (
	"bytes"
	"fmt"
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
	noDups := make([]string, len(check))
	i := 0
	for k := range check {
		noDups[i] = k
		i++
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

func FileExists(filename string) bool {
	info, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, errors.Errorf(errors.InvalidFilePathErrorMsg, path)
	}

	loader := new(properties.Loader)
	loader.Encoding = properties.UTF8
	loader.PreserveFormatting = true

	data, err := os.ReadFile(path)
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
	time := time.Unix(0, timeMs*int64(time.Millisecond))
	return time.UTC().Format("2006-01-02 15:04:05 MST")
}

func ArrayToCommaDelimitedString(arr []string) string {
	size := len(arr)
	switch size {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf(`"%s"`, arr[0])
	case 2:
		return fmt.Sprintf(`"%s" or "%s"`, arr[0], arr[1])
	}

	var delimitedStr strings.Builder
	for _, v := range arr[:size-1] {
		delimitedStr.WriteString(fmt.Sprintf(`"%s", `, v))
	}
	delimitedStr.WriteString(fmt.Sprintf(`or "%s"`, arr[size-1]))

	return delimitedStr.String()
}

func Int32Ptr(x int32) *int32 {
	return &x
}
