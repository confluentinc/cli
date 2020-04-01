package prompt

import (
	"fmt"

	gprompt "github.com/c-bata/go-prompt"
)

type Color256Writer struct {
	gprompt.ConsoleWriter
}

// SetColor sets the text color. Color256Writer will interpret each color
// as a value 0-255, as specified here: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit.
func (w *Color256Writer) SetColor(fg, bg gprompt.Color, bold bool) {
	if bold {
		w.SetDisplayAttributes(fg, bg, gprompt.DisplayBold)
	} else {
		w.SetDisplayAttributes(fg, bg, gprompt.DisplayReset)
	}
	return
}

func NewStdoutColor256Writer() *Color256Writer {
	return &Color256Writer{ConsoleWriter: gprompt.NewStdoutWriter()}
}

// SetDisplayAttributes to set VT100 display attributes.
func (w *Color256Writer) SetDisplayAttributes(fg, bg gprompt.Color, attrs ...gprompt.DisplayAttribute) {
	w.WriteRawStr("\x1b[")        // Control sequence introducer.
	defer w.WriteRaw([]byte{'m'}) // final character
	var separator byte = ';'
	for a := range attrs {
		b := displayAttributeToBytes(gprompt.DisplayAttribute(a))
		w.WriteRaw(b)
		w.WriteRaw([]byte{separator})
	}
	// Begin writing 256 color strings.
	// Foreground.
	w.WriteRawStr("38;5;") // 8-bit foreground escape sequence.
	w.WriteRawStr(fmt.Sprintf("%d", fg))
	// Background.
	w.WriteRaw([]byte{separator})
	w.WriteRawStr("48;5;") // 8-bit background escape sequence.
	w.WriteRawStr(fmt.Sprintf("%d", bg))
	return
}
