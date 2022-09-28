package releasenotes

import (
	"fmt"

	"github.com/confluentinc/cli/internal/pkg/utils"
)

func readTestFile(filePath string) (string, error) {
	fileBytes, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("unable to load output file")
	}
	fileContent := string(fileBytes)
	return utils.NormalizeNewLines(fileContent), nil
}
