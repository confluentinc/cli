package components

import (
	"fmt"
	"strings"

	"github.com/rivo/tview"

	"github.com/confluentinc/cli/internal/pkg/flink/types"
)

// Keyboard shortcuts shown at the bottom.
var shortcuts = []types.Shortcut{
	{KeyText: "Q", Text: "Quit"},
}

func CreateRowView(textView *tview.TextView) *tview.Flex {
	textView.SetDynamicColors(true).SetBorder(true).SetTitle(" Row details ")

	shortcutsView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(true).
		SetWrap(false)
	sb := strings.Builder{}
	for index, shortcut := range shortcuts {
		sb.WriteString(fmt.Sprintf(`[[white]%s] ["%d"][darkcyan]%s[white][""]  `, shortcut.KeyText, index, shortcut.Text))
	}
	shortcutsView.SetText(sb.String())

	return tview.NewFlex().
		SetDirection(tview.FlexRow).
		AddItem(textView, 0, 1, false).
		AddItem(shortcutsView, 1, 1, false)
}
