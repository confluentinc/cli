package components

import (
	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

const ExitRowViewShortcut = "Q"

// Keyboard shortcuts shown at the bottom.
var rowViewShortcuts = []types.Shortcut{
	{KeyText: ExitRowViewShortcut, Text: "Quit"},
}

func CreateRowView(textView *tview.TextView) *tview.Flex {
	textView.SetDynamicColors(true).SetBorder(true).SetTitle(" Row details ")

	shortcuts := NewShortcuts(rowViewShortcuts)

	flex := tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(shortcuts, 1, 1, false)
	return flex
}
