package components

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

func NewShortcuts(shortcuts []types.Shortcut) *tview.TextView {
	shortcutsRef := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	text := formatShortcuts(shortcuts)
	shortcutsRef.SetText(text)

	return shortcutsRef
}

func formatShortcuts(shortcuts []types.Shortcut) string {
	sb := strings.Builder{}
	for index, shortcut := range shortcuts {
		sb.WriteString(fmt.Sprintf(`[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text))
	}
	return sb.String()
}
