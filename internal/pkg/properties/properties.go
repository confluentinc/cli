package properties

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// FileToMap reads key=value pairs from a properties file, ignoring comments and empty lines.
func FileToMap(filename string) (map[string]string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	return toMap(parseLines(string(buf)))
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

// toMap converts a list of key=value strings into a map.
func toMap(configs []string) (map[string]string, error) {
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
