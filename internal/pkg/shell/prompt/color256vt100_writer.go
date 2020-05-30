package prompt

import (
	"fmt"
	"strconv"

	goprompt "github.com/c-bata/go-prompt"
)

type Color256VT100Writer struct {
	goprompt.ConsoleWriter
}

// SetColor sets the text color. Color256VT100Writer will interpret each color
// as a value 0-255, as specified here: https://en.wikipedia.org/wiki/ANSI_escape_code#8-bit.
// To prevent forking of go-prompt, this writer interprets 0 as the default color, not black.
// TODO: To use black... TBD.
func (w *Color256VT100Writer) SetColor(fg, bg goprompt.Color, bold bool) {
	if bold {
		w.setDisplayAttributes(fg, bg, goprompt.DisplayBold)
	} else {
		w.setDisplayAttributes(fg, bg, goprompt.DisplayReset)
	}
	return
}

func NewStdoutColor256VT100Writer() *Color256VT100Writer {
	return &Color256VT100Writer{ConsoleWriter: goprompt.NewStdoutWriter()}
}

// SetDisplayAttributes to set VT100 display attributes.
func (w *Color256VT100Writer) setDisplayAttributes(fg, bg goprompt.Color, attrs ...goprompt.DisplayAttribute) {
	w.WriteRawStr("\x1b[")        // Control sequence introducer.
	defer w.WriteRaw([]byte{'m'}) // final character
	var separator byte = ';'
	for a := range attrs {
		b := displayAttributeToBytes(goprompt.DisplayAttribute(a))
		w.WriteRaw(b)
		w.WriteRaw([]byte{separator})
	}
	// Begin writing 256 color strings.
	// Foreground.
	if fg == 0 {
		w.WriteRawStr("39") // Reset to default fg color. 
	} else {
		w.WriteRawStr(fmt.Sprintf("38;5;%d", fg)) // 8-bit foreground escape sequence.
	}
	w.WriteRaw([]byte{separator})
	// Background.
	if bg == 0 {
		w.WriteRawStr("49") // Reset to default fg color.
	} else {
		w.WriteRawStr(fmt.Sprintf("48;5;%d", bg)) // 8-bit background escape sequence.
	}
}

// displayAttributeToBytes converts a DisplayAttribute to its code in bytes.
func displayAttributeToBytes(attribute goprompt.DisplayAttribute) []byte {
	val := int(attribute)
	s := strconv.Itoa(val)
	return []byte(s)
}
