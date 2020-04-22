package test

import (
	"bytes"
	"github.com/confluentinc/cli/internal/pkg/errors"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
)

// NormalizeNewLines replaces \r\n and \r newline sequences with \n
func NormalizeNewLines(raw string) string {
	normalized := bytes.Replace([]byte(raw), []byte{13, 10}, []byte{10}, -1)
	normalized = bytes.Replace(normalized, []byte{13}, []byte{10}, -1)
	return string(normalized)
}

func WriteConfig(inputConfigFile string, cliName string, destFile string, loginURL string) error {
	input, err := ioutil.ReadFile(inputConfigFile)
	if err != nil {
		return errors.Wrapf(err, "unable to read input config file: %s", inputConfigFile)
	}

	re := regexp.MustCompile("https?://127.0.0.1:[0-9]+")
	inputConfigString := string(input)
	inputConfigString = re.ReplaceAllString(inputConfigString, loginURL)

	err = os.MkdirAll(filepath.Dir(destFile), 0700)
	if err != nil {
		return errors.Wrapf(err, "unable to create config directory: %s", destFile)
	}
	err = ioutil.WriteFile(destFile, []byte(inputConfigString), 0600)
	if err != nil {
		return errors.Wrapf(err, "unable to write config to file: %s", destFile)
	}

	return nil
}
