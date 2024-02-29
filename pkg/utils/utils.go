package utils

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/v3/pkg/errors"
)

const maxFileSize = 1024 * 1024 * 1024 // 1GB

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
	return info != nil && !info.IsDir()
}

func LoadPropertiesFile(path string) (*properties.Properties, error) {
	if !DoesPathExist(path) {
		return nil, fmt.Errorf(errors.InvalidFilePathErrorMsg, path)
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
	normalized := bytes.ReplaceAll(raw, []byte{13, 10}, []byte{10})
	normalized = bytes.ReplaceAll(normalized, []byte{13}, []byte{10})
	return normalized
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

func ArrayToCommaDelimitedString(arr []string, conjunction string) string {
	size := len(arr)
	switch size {
	case 0:
		return ""
	case 1:
		return fmt.Sprintf(`"%s"`, arr[0])
	case 2:
		return fmt.Sprintf(`"%s" %s "%s"`, arr[0], conjunction, arr[1])
	}

	var delimitedStr strings.Builder
	for _, v := range arr[:size-1] {
		delimitedStr.WriteString(fmt.Sprintf(`"%s", `, v))
	}
	delimitedStr.WriteString(fmt.Sprintf(`%s "%s"`, conjunction, arr[size-1]))

	return delimitedStr.String()
}

func Int32Ptr(x int32) *int32 {
	return &x
}

func AddDryRunPrefix(msg string) string {
	return fmt.Sprintf("[DRY RUN] %s", msg)
}

func UploadFile(url, filePath string, formFields map[string]any) error {
	var buffer bytes.Buffer
	writer := multipart.NewWriter(&buffer)

	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	fileSize := fileInfo.Size()
	if fileSize > maxFileSize {
		return fmt.Errorf("File size exceeds the limit of 1GB. Actual size: %d", fileInfo.Size())
	}

	for key, value := range formFields {
		if strValue, ok := value.(string); ok {
			err := writer.WriteField(key, strValue)
			if err != nil {
				return err
			}
		}
	}

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return err
	}
	if _, err := io.Copy(part, file); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	client := &http.Client{
		Timeout: 20 * time.Minute,
	}

	request, err := http.NewRequest("POST", url, &buffer)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	response, err := client.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode >= 400 {
		responseBody, err := ioutil.ReadAll(response.Body)
		return fmt.Errorf("[Response] %s [Error] %v", string(responseBody), err)
	}

	return nil
}
