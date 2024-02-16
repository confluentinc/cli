package properties

import (
	"bytes"
	"fmt"
	"sort"
	"strings"

	"github.com/confluentinc/properties"

	"github.com/confluentinc/cli/v3/pkg/utils"
)

// GetMap reads newline-separated configuration files or comma-separated lists of key=value pairs, and supports configuration values containing commas.
// One known difference compared to Java Property format is it treats the first space as the key value separator.
// For examples:
// key is equal to key=
// key val is equal to key=val
// key val1 val2 is equal to key=val1 val2
// More examples are properties_test.go.
func GetMap(config []string) (map[string]string, error) {
	if len(config) == 1 && utils.FileExists(config[0]) {
		return fileToMap(config[0])
	}

	return ConfigFlagToMap(config)
}

// fileToMap reads key=value pairs from a properties file, ignoring comments and empty lines.
func fileToMap(filename string) (map[string]string, error) {
	prop, err := properties.LoadFile(filename, properties.UTF8)
	if err != nil {
		return nil, err
	}
	return getMapFromProp(prop), err
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

// ConfigFlagToMap reads key=values pairs from the --config flag and supports configuration values containing commas.
func ConfigFlagToMap(configs []string) (map[string]string, error) {
	combinedString := strings.Join(configs, "\n")
	prop, err := properties.LoadString(combinedString)
	if err != nil {
		return nil, err
	}
	return getMapFromProp(prop), err
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

func getMapFromProp(prop *properties.Properties) map[string]string {
	propMap := prop.Map()
	for k, v := range propMap {
		propMap[k] = strings.TrimSpace(v)
	}
	return propMap
}
