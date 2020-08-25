package release_notes

import (
	"fmt"
	"github.com/confluentinc/cli/internal/pkg/utils"
	"io/ioutil"
)

func readTestFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to load output file")
	}
	fileContent := string(fileBytes)
	return utils.NormalizeNewLines(fileContent), nil
}
