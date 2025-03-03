package properties

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/confluentinc/cli/v4/pkg/utils"
)

// GetMap reads newline-separated configuration files or comma-separated lists of key=value pairs, and supports configuration values containing commas.
func GetMap(config []string) (map[string]string, error) {
	if len(config) == 1 && utils.FileExists(config[0]) {
		return fileToMap(config[0])
	}

	return ConfigFlagToMap(config)
}

// fileToMap reads key=value pairs from a properties file, ignoring comments and empty lines.
func fileToMap(filename string) (map[string]string, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ConfigSliceToMap(ParseLines(string(buf)))
}

// ConfigSliceToMap converts a list of key=value strings into a map.
func ConfigSliceToMap(configs []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, config := range configs {
		x := strings.SplitN(config, "=", 2)
		if len(x) < 2 {
			return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, config)
		}

		m[x[0]] = replaceSpecialCharacters(x[1])
	}

	return m, nil
}

func ParseLines(content string) []string {
	var lines []string

	// Support multiline properties
	content = strings.ReplaceAll(content, "\\\n", "")

	for _, line := range strings.Split(content, "\n") {
		line = strings.TrimSpace(line)
		if line != "" && line[0] != '#' {
			lines = append(lines, line)
		}
	}

	return lines
}

// ConfigFlagToMap reads key=values pairs from the --config flag and supports configuration values containing commas.
func ConfigFlagToMap(configs []string) (map[string]string, error) {
	m := make(map[string]string)

	for i := len(configs) - 1; i >= 0; i-- {
		if strings.Contains(configs[i], "=") {
			x := strings.SplitN(configs[i], "=", 2)
			if _, ok := m[x[0]]; !ok {
				m[x[0]] = replaceSpecialCharacters(x[1])
			}
		} else {
			if i-1 >= 0 {
				configs[i-1] += "," + configs[i]
			} else {
				return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, configs[i])
			}
		}
	}

	return m, nil
}

func CreateKeyValuePairs(m map[string]string) string {
	// Sort by keys so the output order is predictable which is helpful for testing.
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	b := new(bytes.Buffer)
	for _, k := range keys {
		fmt.Fprintf(b, "\"%s\"=\"%s\"\n", k, m[k])
	}
	return b.String()
}

func replaceSpecialCharacters(val string) string {
	// Replace \\n, \\r and \\t with newline, carriage return and tab characters as specified in
	// https://docs.oracle.com/cd/E23095_01/Platform.93/ATGProgGuide/html/s0204propertiesfileformat01.html.
	return strings.ReplaceAll(strings.ReplaceAll(
		strings.ReplaceAll(val, "\\n", "\n"), "\\r", "\r"), "\\t", "\t")
}
