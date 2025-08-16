package kafka

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type CopyVariation struct {
	Content string `json:"content"`
	CTA     string `json:"cta"`
}

type CopyScenario struct {
	Variations []CopyVariation `json:"variations"`
}

type CopyData struct {
	Scenarios map[string]CopyScenario `json:"scenarios"`
}

type CopyManager struct {
	data CopyData
}

func NewCopyManager() (*CopyManager, error) {
	// Get the directory of the current file
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return nil, fmt.Errorf("failed to get current file path")
	}
	dir := filepath.Dir(filename)

	// Read the JSON file
	filePath := filepath.Join(dir, "upgrade_suggestions.json")

	file, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read copy data: %w", err)
	}

	var data CopyData
	if err := json.Unmarshal(file, &data); err != nil {
		return nil, fmt.Errorf("failed to parse copy data: %w", err)
	}

	return &CopyManager{data: data}, nil
}

func (cm *CopyManager) GetCopy(scenario string) (string, string, error) {
	scenarioData, exists := cm.data.Scenarios[scenario]
	if !exists {
		return "", "", fmt.Errorf("unknown scenario: %s", scenario)
	}

	variations := scenarioData.Variations
	if len(variations) == 0 {
		return "", "", fmt.Errorf("no variations found for scenario: %s", scenario)
	}

	// For now, simply use the first variation
	// In the future, we could add more sophisticated selection logic if needed
	return variations[0].Content, variations[0].CTA, nil
}

func (cm *CopyManager) FormatCopy(content, cta, id string) string {
	return fmt.Sprintf("%s\n\n%s", content, strings.ReplaceAll(cta, "{{id}}", id))
}
