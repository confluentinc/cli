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
	Weight  int    `json:"weight"`
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
	_, filename, _, _ := runtime.Caller(0)
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

func (cm *CopyManager) GetCopy(scenario string, orgId string) (string, string, error) {
	scenarioData, exists := cm.data.Scenarios[scenario]
	if !exists {
		return "", "", fmt.Errorf("unknown scenario: %s", scenario)
	}

	variations := scenarioData.Variations
	if len(variations) == 0 {
		return "", "", fmt.Errorf("no variations found for scenario: %s", scenario)
	}

	// Calculate total weight
	totalWeight := 0
	for _, v := range variations {
		totalWeight += v.Weight
	}

	// Use orgId to select variation (simple modulo for now)
	// This can be replaced with LaunchDarkly logic later
	hash := 0
	for _, c := range orgId {
		hash = (hash + int(c)) % totalWeight
	}

	// Select variation based on hash
	currentWeight := 0
	for _, v := range variations {
		currentWeight += v.Weight
		if hash < currentWeight {
			return v.Content, v.CTA, nil
		}
	}

	// Fallback to first variation
	return variations[0].Content, variations[0].CTA, nil
}

func (cm *CopyManager) FormatCopy(content, cta, id string) string {
	formattedContent := content
	formattedCTA := strings.ReplaceAll(cta, "{{id}}", id)
	return formattedContent + "\n\n" + formattedCTA
}
