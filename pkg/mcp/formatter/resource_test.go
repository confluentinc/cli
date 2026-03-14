package formatter

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Additional tests for extractID edge cases not covered in json_test.go
func TestExtractID_EdgeCases(t *testing.T) {
	t.Run("id field is not a string", func(t *testing.T) {
		item := map[string]interface{}{
			"id": 12345,
		}
		result := extractID(item)
		assert.Equal(t, "", result)
	})

	t.Run("empty id field", func(t *testing.T) {
		item := map[string]interface{}{
			"id": "",
		}
		result := extractID(item)
		assert.Equal(t, "", result)
	})
}

// Additional tests for extractName edge cases not covered in json_test.go
func TestExtractName_EdgeCases(t *testing.T) {
	t.Run("display_name has priority over label", func(t *testing.T) {
		item := map[string]interface{}{
			"id":           "lkc-123",
			"display_name": "Display Name",
			"label":        "Label",
		}
		result := extractName(item)
		assert.Equal(t, "Display Name", result)
	})

	t.Run("label has priority over title", func(t *testing.T) {
		item := map[string]interface{}{
			"id":    "lkc-123",
			"label": "Label",
			"title": "Title",
		}
		result := extractName(item)
		assert.Equal(t, "Label", result)
	})

	t.Run("name field is not a string", func(t *testing.T) {
		item := map[string]interface{}{
			"id":   "lkc-123",
			"name": 12345,
		}
		result := extractName(item)
		assert.Equal(t, "", result)
	})

	t.Run("empty name field is skipped", func(t *testing.T) {
		item := map[string]interface{}{
			"id":           "lkc-123",
			"name":         "",
			"display_name": "Display Name",
		}
		result := extractName(item)
		assert.Equal(t, "Display Name", result)
	})
}
