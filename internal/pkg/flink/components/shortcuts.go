package components

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

// Keyboard shortcuts shown at the bottom.
var appShortcuts = []types.Shortcut{
	{KeyText: "Q", Text: "Quit"},
	{KeyText: "M", Text: "Toggle Result Mode"},
	{KeyText: "P", Text: "Toggle Auto Refresh"},
	{KeyText: "R", Text: "Live results"},
	{KeyText: "H/L", Text: "Fast scroll ▲/▼"},
}

func Shortcuts() *tview.TextView {
	shortcutsRef := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)

	shortcutsRef.SetText(formatShortcuts(appShortcuts))

	return shortcutsRef
}

func formatShortcuts(appShortcuts []types.Shortcut) string {
	sb := strings.Builder{}
	for index, shortcut := range appShortcuts {
		sb.WriteString(fmt.Sprintf(`[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text))
	}
	return sb.String()
}
