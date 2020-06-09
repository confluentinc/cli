package local

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

var confluentControlCenter = "share/java/confluent-control-center/control-center-*.jar"

func findConfluentFile(pattern string) ([]string, error) {
	confluentHome, err := getConfluentHome()
	if err != nil {
		return []string{}, err
	}

	path := filepath.Join(confluentHome, pattern)
	matches, err := filepath.Glob(path)
	if err != nil {
		return []string{}, err
	}

	for i := range matches {
		matches[i], err = filepath.Rel(confluentHome, matches[i])
		if err != nil {
			return []string{}, err
		}
	}
	return matches, nil
}

func getConfluentHome() (string, error) {
	confluentHome := os.Getenv("CONFLUENT_HOME")
	if confluentHome == "" {
		return "", fmt.Errorf("set environment variable CONFLUENT_HOME")
	}
	return confluentHome, nil
}

func isConfluentPlatform() (bool, error) {
	files, err := findConfluentFile(confluentControlCenter)
	if err != nil {
		return false, err
	}
	return len(files) > 0, nil
}

func buildTabbedList(slice []string) string {
	sort.Strings(slice)

	var list strings.Builder
	for _, x := range slice {
		fmt.Fprintf(&list, "  %s\n", x)
	}
	return list.String()
}
