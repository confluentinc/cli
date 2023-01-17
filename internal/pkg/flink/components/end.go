package components

import (
	"fmt"

	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

// End shows the final slide.
func End() tview.Primitive {
	textView := tview.NewTextView().SetDoneFunc(func(key tcell.Key) {
	})
	url := "https://github.com/rivo/tview"
	fmt.Fprint(textView, url)
	return Center(len(url), 1, textView)
}
