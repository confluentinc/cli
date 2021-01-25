package release_notes

import (
	"fmt"
	"io/ioutil"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

func readTestFile(filePath string) (string, error) {
	fileBytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to load output file")
	}
	fileContent := string(fileBytes)
	return utils.NormalizeNewLines(fileContent), nil
}
