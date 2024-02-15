package properties

import (
	"bytes"
	"fmt"
	"github.com/confluentinc/properties"
	"os"
	"sort"
	"strings"

	"github.com/confluentinc/cli/v3/pkg/utils"
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

	return ConfigSliceToMap(parseLines(string(buf)))
}

// ConfigSliceToMap converts a list of key=value strings into a map.
func ConfigSliceToMap(configs []string) (map[string]string, error) {
	m := make(map[string]string)

	for _, config := range configs {
		x := strings.SplitN(config, "=", 2)
		if len(x) < 2 {
			return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, config)
		}

		m[x[0]] = x[1]
	}

	return m, nil
}

func parseLines(content string) []string {
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
				m[x[0]] = x[1]
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

// GetMapWithJavaPropertyParsing reads key=value pairs from the file or string slices, according to Java property
// format as close as possible. One known difference is it treats the first space as the key value separator.
// For examples:
// key is equal to key=
// key val is equal to key=val
// key val1 val2 is equal to key=val1 val2
// More examples are properties_test.go.
func GetMapWithJavaPropertyParsing(config []string) (map[string]string, error) {
	if len(config) == 1 && utils.FileExists(config[0]) {
		return fileToMapWithJavaPropertyParsing(config[0])
	}
	return ConfigFlagToMapWithJavaPropertyParsing(config)
}

// fileToMapWithJavaPropertyParsing reads key=value pairs from a properties file, according to Java property file
// format as close as possible.
func fileToMapWithJavaPropertyParsing(filename string) (map[string]string, error) {
	prop, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return prop.Map(), err
}

// ConfigFlagToMapWithJavaPropertyParsing reads key=value pairs from the string slices, according to Java property
// format as close as possible.
func ConfigFlagToMapWithJavaPropertyParsing(configs []string) (map[string]string, error) {
	combinedString := strings.Join(configs, "\n")
	prop, err := properties.LoadString(combinedString)
	if err != nil {
		return nil, err
	}
	return prop.Map(), err
}
