package release_notes

import (
	"fmt"
	"io/ioutil"

	testUtils "github.com/confluentinc/cli/test"
)

func readTestFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to load output file")
	}
	fileContent := string(fileBytes)
	return testUtils.NormalizeNewLines(fileContent), nil
}
