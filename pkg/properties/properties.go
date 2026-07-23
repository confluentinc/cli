package properties

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
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
func fileToMap(filename string, rawValueKeys ...string) (map[string]string, error) {
	buf, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return ConfigSliceToMap(ParseLines(string(buf)), rawValueKeys...)
}

// ConfigSliceToMap converts a list of key=value strings into a map.
func ConfigSliceToMap(configs []string, rawValueKeys ...string) (map[string]string, error) {
	m := make(map[string]string)

	for _, config := range configs {
		x := strings.SplitN(config, "=", 2)
		if len(x) < 2 {
			return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, config)
		}

		// rawValueKeys are stored as-is, all other values get un-escaped.
		if slices.Contains(rawValueKeys, x[0]) {
			m[x[0]] = x[1]
		} else {
			m[x[0]] = replaceSpecialCharacters(x[1])
		}
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

// GetMapFromArray reads configuration from a configuration file or from a StringArray. It supports values containing commas.
// Values whose keys are listed in jsonValueKeys are treated as JSON.
// Use this instead of GetMap for flags registered with cmd.AddTopicConfigFlag.
func GetMapFromArray(config []string, jsonValueKeys ...string) (map[string]string, error) {
	if len(config) == 1 && utils.FileExists(config[0]) {
		return fileToMap(config[0], jsonValueKeys...)
	}

	return configArrayToMap(config, jsonValueKeys...)
}

// configArrayToMap parses raw config elements into a map, each element is split on commas into "key=value" pairs.
// JSON config values as indicated by jsonValueKeys are preserved and validated.
func configArrayToMap(configs []string, jsonValueKeys ...string) (map[string]string, error) {
	m := make(map[string]string)

	for _, config := range configs {
		for _, pair := range splitConfigPairs(config, jsonValueKeys) {
			x := strings.SplitN(pair, "=", 2)
			if len(x) < 2 {
				return nil, fmt.Errorf(`failed to parse "key=value" pattern from configuration: %s`, pair)
			}

			if slices.Contains(jsonValueKeys, x[0]) {
				if !json.Valid([]byte(strings.TrimSpace(x[1]))) {
					return nil, fmt.Errorf(`failed to parse JSON value for configuration "%s": %s`, x[0], x[1])
				}
				m[x[0]] = x[1]
			} else {
				m[x[0]] = replaceSpecialCharacters(x[1])
			}
		}
	}

	return m, nil
}

// splitConfigPairs splits config into raw "key=value" pair strings, honoring
// comma-separated values and JSON values for jsonValueKeys.
func splitConfigPairs(config string, jsonValueKeys []string) []string {
	var pairs []string

	current := ""
	for _, fragment := range strings.Split(config, ",") {
		switch {
		case current == "":
			current = fragment
		case strings.Contains(fragment, "=") && pairComplete(current, jsonValueKeys):
			pairs = append(pairs, current)
			current = fragment
		default:
			// A comma inside the current value: glue the fragment back on.
			current += "," + fragment
		}
	}
	if current != "" {
		pairs = append(pairs, current)
	}

	return pairs
}

// pairComplete reports whether an accumulated "key=value" is complete. For a jsonValueKey the value must be valid JSON.
func pairComplete(pair string, jsonValueKeys []string) bool {
	x := strings.SplitN(pair, "=", 2)
	if len(x) < 2 {
		return true
	}

	if slices.Contains(jsonValueKeys, x[0]) {
		return json.Valid([]byte(strings.TrimSpace(x[1])))
	}

	return true
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
