package properties

import (
	"fmt"
	"io/ioutil"
	"strings"
)

// FileToMap reads the key=value pairs from a properties file, ignoring comments and empty lines.
func FileToMap(filename string) (map[string]string, error) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	lines := parseLines(string(buf))
	return ToMap(lines)
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

// ToMap converts a list of key=value strings into a map.
func ToMap(configs []string) (map[string]string, error) {
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
