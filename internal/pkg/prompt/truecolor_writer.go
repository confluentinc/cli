package prompt

import (
	"fmt"
	"strconv"

	gprompt "github.com/c-bata/go-prompt"
)

type TrueColorWriter struct {
	gprompt.ConsoleWriter
}

func (w *TrueColorWriter) SetColor(fg, bg gprompt.Color, bold bool) {
	if bold {
		w.SetDisplayAttributes(fg, bg, gprompt.DisplayBold)
	} else {
		w.SetDisplayAttributes(fg, bg, gprompt.DisplayReset)
	}
	return
}

func NewStdoutTrueColorWriter() *TrueColorWriter {
	return &TrueColorWriter{ConsoleWriter: gprompt.NewStdoutWriter()}
}

// SetDisplayAttributes to set VT100 display attributes.
// Handle separator bytes. 
func (w *TrueColorWriter) SetDisplayAttributes(fg, bg gprompt.Color, attrs ...gprompt.DisplayAttribute) {
	w.WriteRawStr("\x1b[")        // Control sequence introducer.
	defer w.WriteRaw([]byte{'m'}) // final character
	var separator byte = ';'
	for a := range attrs {
		b := displayAttributeToBytes(gprompt.DisplayAttribute(a))
		w.WriteRaw(b)
		w.WriteRaw([]byte{separator})
	}
	// Begin writing true color strings.
	// Foreground.
	w.WriteRawStr("38;2;") // true color foreground code. 
	r, g, b := hexToRGB(int(fg))
	w.WriteRawStr(fmt.Sprintf("%d;%d;%d", r, g, b))
	// Background.
	w.WriteRaw([]byte{separator})
	w.WriteRawStr("48;2;") // true color background code. 
	r, g, b = hexToRGB(int(bg))
	w.WriteRawStr(fmt.Sprintf("%d;%d;%d", r, g, b))
	return
}

func hexToRGB(hexCode int) (r, g, b int) {
	r = (hexCode >> 16) & 0xFF
	g = (hexCode >> 8) & 0xFF
	b = hexCode & 0xFF
	return
}

// displayAttributeToBytes converts a DisplayAttribute to its code in bytes.
func displayAttributeToBytes(attribute gprompt.DisplayAttribute) []byte {
	val := int(attribute)
	s := strconv.Itoa(val)
	return []byte(s)
}
